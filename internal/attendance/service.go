package attendance

import (
"errors"
"time"

"github.com/beellz/kumbhpos/pkg/db"
)

type Record struct {
ID          string     `db:"id"          json:"id"`
StaffID     string     `db:"staff_id"    json:"staff_id"`
StaffName   string     `db:"staff_name"  json:"staff_name"`
ShopID      string     `db:"shop_id"     json:"shop_id"`
ShopName    string     `db:"shop_name"   json:"shop_name"`
Role        string     `db:"role"        json:"role"`
ClockIn     time.Time  `db:"clock_in"    json:"clock_in"`
ClockOut    *time.Time `db:"clock_out"   json:"clock_out"`
ShiftNotes  string     `db:"shift_notes" json:"shift_notes"`
HoursWorked float64    `json:"hours_worked"`
}

type ClockInRequest struct {
StaffID    string `json:"staff_id"   binding:"required"`
ShopID     string `json:"shop_id"    binding:"required"`
}

type ClockOutRequest struct {
AttendanceID string `json:"attendance_id" binding:"required"`
ShiftNotes   string `json:"shift_notes"`
}

func ClockIn(req ClockInRequest) (Record, error) {
// Check if already clocked in
var count int
_ = db.DB.Get(&count,
`SELECT COUNT(*) FROM attendance
 WHERE staff_id=$1 AND clock_out IS NULL`,
req.StaffID)
if count > 0 {
return Record{}, errors.New("already clocked in — please clock out first")
}

var rec Record
err := db.DB.Get(&rec, `
INSERT INTO attendance (staff_id, shop_id)
VALUES ($1, $2)
RETURNING id, staff_id, shop_id, clock_in,
          clock_out, shift_notes`,
req.StaffID, req.ShopID)
if err != nil {
return Record{}, err
}

// Hydrate names
_ = db.DB.Get(&rec.StaffName,
`SELECT name FROM staff WHERE id=$1`, rec.StaffID)
_ = db.DB.Get(&rec.ShopName,
`SELECT name FROM shops WHERE id=$1`, rec.ShopID)
return rec, nil
}

func ClockOut(req ClockOutRequest) (Record, error) {
var rec Record
err := db.DB.Get(&rec, `
UPDATE attendance
SET clock_out   = NOW(),
    shift_notes = $2
WHERE id = $1 AND clock_out IS NULL
RETURNING id, staff_id, shop_id, clock_in,
          clock_out, shift_notes`,
req.AttendanceID, req.ShiftNotes)
if err != nil {
return Record{}, errors.New("record not found or already clocked out")
}
_ = db.DB.Get(&rec.StaffName,
`SELECT name FROM staff WHERE id=$1`, rec.StaffID)
_ = db.DB.Get(&rec.ShopName,
`SELECT name FROM shops WHERE id=$1`, rec.ShopID)
if rec.ClockOut != nil {
rec.HoursWorked = rec.ClockOut.Sub(rec.ClockIn).Hours()
}
return rec, nil
}

func GetAttendance(shopID, date string) ([]Record, error) {
query := `
SELECT
a.id,
a.staff_id,
st.name  AS staff_name,
a.shop_id,
s.name   AS shop_name,
st.role,
a.clock_in,
a.clock_out,
COALESCE(a.shift_notes, '') AS shift_notes
FROM attendance a
JOIN staff st ON st.id = a.staff_id
JOIN shops  s  ON s.id  = a.shop_id
WHERE 1=1`

args := []interface{}{}
idx  := 1
if shopID != "" {
query += ` AND a.shop_id = $` + itoa(idx)
args   = append(args, shopID)
idx++
}
if date != "" {
query += ` AND DATE(a.clock_in) = $` + itoa(idx)
args   = append(args, date)
idx++
}
query += ` ORDER BY a.clock_in DESC`

var list []Record
err := db.DB.Select(&list, query, args...)
for i := range list {
if list[i].ClockOut != nil {
list[i].HoursWorked =
list[i].ClockOut.Sub(list[i].ClockIn).Hours()
}
}
return list, err
}

func GetActiveShifts(shopID string) ([]Record, error) {
query := `
SELECT
a.id,
a.staff_id,
st.name  AS staff_name,
a.shop_id,
s.name   AS shop_name,
st.role,
a.clock_in,
a.clock_out,
COALESCE(a.shift_notes, '') AS shift_notes
FROM attendance a
JOIN staff st ON st.id = a.staff_id
JOIN shops  s  ON s.id  = a.shop_id
WHERE a.clock_out IS NULL`

args := []interface{}{}
if shopID != "" {
query += ` AND a.shop_id = $1`
args   = append(args, shopID)
}
query += ` ORDER BY a.clock_in ASC`

var list []Record
err := db.DB.Select(&list, query, args...)
return list, err
}

func GetStaffStatus(staffID string) (*Record, error) {
var rec Record
err := db.DB.Get(&rec, `
SELECT
a.id,
a.staff_id,
st.name  AS staff_name,
a.shop_id,
s.name   AS shop_name,
st.role,
a.clock_in,
a.clock_out,
COALESCE(a.shift_notes, '') AS shift_notes
FROM attendance a
JOIN staff st ON st.id = a.staff_id
JOIN shops  s  ON s.id  = a.shop_id
WHERE a.staff_id = $1 AND a.clock_out IS NULL
LIMIT 1`,
staffID)
if err != nil {
return nil, nil // not clocked in — not an error
}
return &rec, nil
}

func itoa(i int) string {
return string(rune('0' + i))
}
