package admin

import (
"github.com/beellz/kumbhpos/pkg/db"
"golang.org/x/crypto/bcrypt"
)

// ── Shop management ──────────────────────────────────────────

type ShopDetail struct {
ID       string `db:"id"        json:"id"`
Name     string `db:"name"      json:"name"`
Location string `db:"location"  json:"location"`
IsActive bool   `db:"is_active" json:"is_active"`
TxnCount int    `db:"txn_count" json:"txn_count"`
TotalGMV float64 `db:"total_gmv" json:"total_gmv"`
}

func ListShopsDetailed() ([]ShopDetail, error) {
var shops []ShopDetail
err := db.DB.Select(&shops, `
SELECT
s.id, s.name, s.location, s.is_active,
COUNT(t.id)                        AS txn_count,
COALESCE(SUM(t.total_amount), 0)   AS total_gmv
FROM shops s
LEFT JOIN transactions t
ON t.shop_id = s.id AND t.payment_status = 'confirmed'
GROUP BY s.id
ORDER BY total_gmv DESC
`)
return shops, err
}

func SetShopActive(shopID string, active bool) error {
_, err := db.DB.Exec(
`UPDATE shops SET is_active=$1 WHERE id=$2`,
active, shopID)
return err
}

// ── Staff management ─────────────────────────────────────────

type StaffDetail struct {
ID       string `db:"id"        json:"id"`
ShopID   string `db:"shop_id"   json:"shop_id"`
ShopName string `db:"shop_name" json:"shop_name"`
Name     string `db:"name"      json:"name"`
Role     string `db:"role"      json:"role"`
IsActive bool   `db:"is_active" json:"is_active"`
TxnCount int    `db:"txn_count" json:"txn_count"`
}

func ListAllStaff() ([]StaffDetail, error) {
var list []StaffDetail
err := db.DB.Select(&list, `
SELECT
st.id, st.shop_id, s.name AS shop_name,
st.name, st.role, st.is_active,
COUNT(t.id) AS txn_count
FROM staff st
JOIN shops s ON s.id = st.shop_id
LEFT JOIN transactions t ON t.staff_id = st.id
GROUP BY st.id, s.name
ORDER BY s.name, st.name
`)
return list, err
}

func ResetStaffPIN(staffID, newPIN string) error {
hash, err := bcrypt.GenerateFromPassword([]byte(newPIN), bcrypt.DefaultCost)
if err != nil {
return err
}
_, err = db.DB.Exec(
`UPDATE staff SET pin_hash=$1 WHERE id=$2`,
string(hash), staffID)
return err
}

func SetStaffActive(staffID string, active bool) error {
_, err := db.DB.Exec(
`UPDATE staff SET is_active=$1 WHERE id=$2`,
active, staffID)
return err
}

func ChangeStaffRole(staffID, role string) error {
_, err := db.DB.Exec(
`UPDATE staff SET role=$1 WHERE id=$2`,
role, staffID)
return err
}

// ── Item management ───────────────────────────────────────────

type ItemDetail struct {
ID         string  `db:"id"          json:"id"`
Name       string  `db:"name"        json:"name"`
NameHi     string  `db:"name_hi"     json:"name_hi"`
Price      float64 `db:"price"       json:"price"`
TaxRate    float64 `db:"tax_rate"    json:"tax_rate"`
CategoryID string  `db:"category_id" json:"category_id"`
CatName    string  `db:"cat_name"    json:"category_name"`
IsActive   bool    `db:"is_active"   json:"is_active"`
}

func ListAllItems() ([]ItemDetail, error) {
var items []ItemDetail
err := db.DB.Select(&items, `
SELECT
i.id, i.name, i.name_hi, i.price, i.tax_rate,
i.category_id, c.name AS cat_name, i.is_active
FROM items i
JOIN categories c ON c.id = i.category_id
ORDER BY c.name, i.name
`)
return items, err
}

func UpdateItemPrice(itemID string, price float64) error {
_, err := db.DB.Exec(
`UPDATE items SET price=$1 WHERE id=$2`,
price, itemID)
return err
}

func SetItemActive(itemID string, active bool) error {
_, err := db.DB.Exec(
`UPDATE items SET is_active=$1 WHERE id=$2`,
active, itemID)
return err
}

// ── Global stats ──────────────────────────────────────────────

type GlobalStats struct {
TotalShops    int     `db:"total_shops"    json:"total_shops"`
ActiveShops   int     `db:"active_shops"   json:"active_shops"`
TotalStaff    int     `db:"total_staff"    json:"total_staff"`
TotalItems    int     `db:"total_items"    json:"total_items"`
TotalGMVAllTime float64 `db:"total_gmv"   json:"total_gmv_all_time"`
TotalTxns     int     `db:"total_txns"     json:"total_txns"`
BeellzAllTime float64 `db:"beellz_share"   json:"beellz_share_all_time"`
TodayGMV      float64 `db:"today_gmv"      json:"today_gmv"`
TodayTxns     int     `db:"today_txns"     json:"today_txns"`
}

func GetGlobalStats() (GlobalStats, error) {
var stats GlobalStats
err := db.DB.Get(&stats, `
SELECT
(SELECT COUNT(*) FROM shops)                        AS total_shops,
(SELECT COUNT(*) FROM shops WHERE is_active=true)   AS active_shops,
(SELECT COUNT(*) FROM staff  WHERE is_active=true)  AS total_staff,
(SELECT COUNT(*) FROM items  WHERE is_active=true)  AS total_items,
COALESCE((SELECT SUM(total_amount) FROM transactions
          WHERE payment_status='confirmed'), 0)      AS total_gmv,
(SELECT COUNT(*) FROM transactions
 WHERE payment_status='confirmed')                   AS total_txns,
COALESCE((SELECT SUM(total_amount)*0.4 FROM transactions
          WHERE payment_status='confirmed'), 0)      AS beellz_share,
COALESCE((SELECT SUM(total_amount) FROM transactions
          WHERE payment_status='confirmed'
            AND created_at >= CURRENT_DATE), 0)      AS today_gmv,
(SELECT COUNT(*) FROM transactions
 WHERE payment_status='confirmed'
   AND created_at >= CURRENT_DATE)                   AS today_txns
`)
return stats, err
}
