package staff

import (
"github.com/beellz/kumbhpos/pkg/db"
"golang.org/x/crypto/bcrypt"
)

type Staff struct {
ID      string `db:"id"      json:"id"`
ShopID  string `db:"shop_id" json:"shop_id"`
Name    string `db:"name"    json:"name"`
Role    string `db:"role"    json:"role"`
IsActive bool  `db:"is_active" json:"is_active"`
}

type CreateStaffRequest struct {
ShopID string `json:"shop_id" binding:"required"`
Name   string `json:"name"    binding:"required"`
PIN    string `json:"pin"     binding:"required"`
Role   string `json:"role"`
}

func CreateStaff(req CreateStaffRequest) (Staff, error) {
if req.Role == "" {
req.Role = "cashier"
}
hash, err := bcrypt.GenerateFromPassword([]byte(req.PIN), bcrypt.DefaultCost)
if err != nil {
return Staff{}, err
}
var s Staff
err = db.DB.QueryRowx(
`INSERT INTO staff (shop_id, name, pin_hash, role)
 VALUES ($1, $2, $3, $4)
 RETURNING id, shop_id, name, role, is_active`,
req.ShopID, req.Name, string(hash), req.Role,
).StructScan(&s)
return s, err
}

func ListStaffByShop(shopID string) ([]Staff, error) {
var list []Staff
err := db.DB.Select(&list,
`SELECT id, shop_id, name, role, is_active FROM staff WHERE shop_id=$1`,
shopID)
return list, err
}
