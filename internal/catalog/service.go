package catalog

import (
"github.com/beellz/kumbhpos/pkg/db"
)

type Category struct {
ID     string `db:"id"      json:"id"`
Name   string `db:"name"    json:"name"`
NameHi string `db:"name_hi" json:"name_hi"`
}

type Item struct {
ID         string  `db:"id"          json:"id"`
CategoryID string  `db:"category_id" json:"category_id"`
Name       string  `db:"name"        json:"name"`
NameHi     string  `db:"name_hi"     json:"name_hi"`
Price      float64 `db:"price"       json:"price"`
TaxRate    float64 `db:"tax_rate"    json:"tax_rate"`
IsActive   bool    `db:"is_active"   json:"is_active"`
}

type CreateCategoryRequest struct {
Name   string `json:"name"    binding:"required"`
NameHi string `json:"name_hi"`
}

type CreateItemRequest struct {
CategoryID string  `json:"category_id" binding:"required"`
Name       string  `json:"name"        binding:"required"`
NameHi     string  `json:"name_hi"`
Price      float64 `json:"price"       binding:"required"`
TaxRate    float64 `json:"tax_rate"`
}

type AssignItemRequest struct {
ItemID           string `json:"item_id"            binding:"required"`
StockQty         int    `json:"stock_qty"`
LowStockThreshold int   `json:"low_stock_threshold"`
}

func CreateCategory(req CreateCategoryRequest) (Category, error) {
var cat Category
err := db.DB.QueryRowx(
`INSERT INTO categories (name, name_hi) VALUES ($1, $2)
 RETURNING id, name, name_hi`,
req.Name, req.NameHi,
).StructScan(&cat)
return cat, err
}

func ListCategories() ([]Category, error) {
var cats []Category
err := db.DB.Select(&cats, `SELECT id, name, name_hi FROM categories ORDER BY name`)
return cats, err
}

func CreateItem(req CreateItemRequest) (Item, error) {
var item Item
err := db.DB.QueryRowx(
`INSERT INTO items (category_id, name, name_hi, price, tax_rate)
 VALUES ($1, $2, $3, $4, $5)
 RETURNING id, category_id, name, name_hi, price, tax_rate, is_active`,
req.CategoryID, req.Name, req.NameHi, req.Price, req.TaxRate,
).StructScan(&item)
return item, err
}

func ListItems() ([]Item, error) {
var items []Item
err := db.DB.Select(&items,
`SELECT id, category_id, name, name_hi, price, tax_rate, is_active
 FROM items WHERE is_active=true ORDER BY name`)
return items, err
}

func AssignItemToShop(shopID string, req AssignItemRequest) error {
threshold := req.LowStockThreshold
if threshold == 0 {
threshold = 10
}
_, err := db.DB.Exec(
`INSERT INTO shop_items (shop_id, item_id, stock_qty, low_stock_threshold)
 VALUES ($1, $2, $3, $4)
 ON CONFLICT (shop_id, item_id) DO UPDATE
 SET stock_qty = $3, low_stock_threshold = $4`,
shopID, req.ItemID, req.StockQty, threshold,
)
return err
}

func GetShopCatalog(shopID string) ([]Item, error) {
var items []Item
err := db.DB.Select(&items,
`SELECT i.id, i.category_id, i.name, i.name_hi, i.price, i.tax_rate, i.is_active
 FROM items i
 JOIN shop_items si ON si.item_id = i.id
 WHERE si.shop_id = $1 AND i.is_active = true
 ORDER BY i.name`,
shopID)
return items, err
}
