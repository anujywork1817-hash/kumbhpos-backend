package inventory

import (
"github.com/beellz/kumbhpos/internal/dashboard"
"github.com/beellz/kumbhpos/pkg/db"
)

type StockLevel struct {
ShopID           string  `db:"shop_id"            json:"shop_id"`
ShopName         string  `db:"shop_name"          json:"shop_name"`
ItemID           string  `db:"item_id"            json:"item_id"`
ItemName         string  `db:"item_name"          json:"item_name"`
StockQty         int     `db:"stock_qty"          json:"stock_qty"`
LowStockThreshold int    `db:"low_stock_threshold" json:"low_stock_threshold"`
IsLow            bool    `db:"is_low"             json:"is_low"`
}

type RestockRequest struct {
ShopID      string `db:"shop_id"      json:"shop_id"`
ShopName    string `db:"shop_name"    json:"shop_name"`
ItemID      string `db:"item_id"      json:"item_id"`
ItemName    string `db:"item_name"    json:"item_name"`
RequestedQty int   `db:"requested_qty" json:"requested_qty"`
Status      string `db:"status"       json:"status"`
ID          string `db:"id"           json:"id"`
}

type AdjustStockRequest struct {
ShopID  string `json:"shop_id"  binding:"required"`
ItemID  string `json:"item_id"  binding:"required"`
Qty     int    `json:"qty"      binding:"required"`
Reason  string `json:"reason"`
}

type CreateRestockRequest struct {
ShopID       string `json:"shop_id"        binding:"required"`
ItemID       string `json:"item_id"        binding:"required"`
RequestedQty int    `json:"requested_qty"  binding:"required"`
}

func GetStockLevels(shopID string) ([]StockLevel, error) {
query := `
SELECT
si.shop_id,
s.name                              AS shop_name,
si.item_id,
i.name                              AS item_name,
si.stock_qty,
si.low_stock_threshold,
(si.stock_qty <= si.low_stock_threshold) AS is_low
FROM shop_items si
JOIN shops s ON s.id = si.shop_id
JOIN items i ON i.id = si.item_id
WHERE 1=1`

var args []interface{}
if shopID != "" {
query += " AND si.shop_id = $1"
args = append(args, shopID)
}
query += " ORDER BY is_low DESC, i.name"

var levels []StockLevel
err := db.DB.Select(&levels, query, args...)
return levels, err
}

func GetLowStockAlerts() ([]StockLevel, error) {
var levels []StockLevel
err := db.DB.Select(&levels, `
SELECT
si.shop_id,
s.name  AS shop_name,
si.item_id,
i.name  AS item_name,
si.stock_qty,
si.low_stock_threshold,
true    AS is_low
FROM shop_items si
JOIN shops s ON s.id = si.shop_id
JOIN items i ON i.id = si.item_id
WHERE si.stock_qty <= si.low_stock_threshold
ORDER BY si.stock_qty ASC
`)
return levels, err
}

func AdjustStock(req AdjustStockRequest) error {
_, err := db.DB.Exec(`
UPDATE shop_items SET stock_qty = stock_qty + $1
WHERE shop_id = $2 AND item_id = $3`,
req.Qty, req.ShopID, req.ItemID,
)
if err != nil {
return err
}
// Broadcast stock update
go dashboard.PushLiveUpdate("stock_adjusted", map[string]interface{}{
"shop_id": req.ShopID,
"item_id": req.ItemID,
"qty":     req.Qty,
"reason":  req.Reason,
})
return nil
}

func SubmitRestockRequest(req CreateRestockRequest) error {
_, err := db.DB.Exec(`
INSERT INTO restock_requests (shop_id, item_id, requested_qty)
VALUES ($1, $2, $3)`,
req.ShopID, req.ItemID, req.RequestedQty,
)
if err != nil {
return err
}
go dashboard.PushLiveUpdate("restock_requested", map[string]interface{}{
"shop_id": req.ShopID,
"item_id": req.ItemID,
"qty":     req.RequestedQty,
})
return nil
}

func GetRestockRequests(status string) ([]RestockRequest, error) {
query := `
SELECT
rr.id,
rr.shop_id,
s.name  AS shop_name,
rr.item_id,
i.name  AS item_name,
rr.requested_qty,
rr.status
FROM restock_requests rr
JOIN shops s ON s.id = rr.shop_id
JOIN items i ON i.id = rr.item_id`

var args []interface{}
if status != "" {
query += " WHERE rr.status = $1"
args = append(args, status)
}
query += " ORDER BY rr.created_at DESC"

var list []RestockRequest
err := db.DB.Select(&list, query, args...)
return list, err
}

func ApproveRestockRequest(requestID string) error {
// Get the request details
var req struct {
ShopID       string `db:"shop_id"`
ItemID       string `db:"item_id"`
RequestedQty int    `db:"requested_qty"`
}
err := db.DB.Get(&req,
`SELECT shop_id, item_id, requested_qty FROM restock_requests WHERE id=$1`,
requestID)
if err != nil {
return err
}

// Update stock
_, err = db.DB.Exec(`
UPDATE shop_items SET stock_qty = stock_qty + $1
WHERE shop_id=$2 AND item_id=$3`,
req.RequestedQty, req.ShopID, req.ItemID,
)
if err != nil {
return err
}

// Mark request fulfilled
_, err = db.DB.Exec(
`UPDATE restock_requests SET status='fulfilled' WHERE id=$1`,
requestID)
return err
}

func TransferStock(fromShopID, toShopID, itemID string, qty int) error {
tx, err := db.DB.Beginx()
if err != nil {
return err
}
defer tx.Rollback()

// Deduct from source
_, err = tx.Exec(`
UPDATE shop_items SET stock_qty = stock_qty - $1
WHERE shop_id=$2 AND item_id=$3 AND stock_qty >= $1`,
qty, fromShopID, itemID)
if err != nil {
return err
}

// Add to destination
_, err = tx.Exec(`
UPDATE shop_items SET stock_qty = stock_qty + $1
WHERE shop_id=$2 AND item_id=$3`,
qty, toShopID, itemID)
if err != nil {
return err
}

return tx.Commit()
}
