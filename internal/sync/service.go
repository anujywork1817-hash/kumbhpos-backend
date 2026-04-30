package sync

import (
"time"
"github.com/beellz/kumbhpos/internal/dashboard"
"github.com/beellz/kumbhpos/pkg/db"
)

type TerminalTransaction struct {
TerminalTxnID  string      `json:"terminal_txn_id"  binding:"required"`
ShopID         string      `json:"shop_id"          binding:"required"`
StaffID        string      `json:"staff_id"         binding:"required"`
PaymentMode    string      `json:"payment_mode"     binding:"required"`
TotalAmount    float64     `json:"total_amount"     binding:"required"`
DiscountAmount float64     `json:"discount_amount"`
CashReceived   float64     `json:"cash_received"`
ChangeGiven    float64     `json:"change_given"`
CreatedAt      string      `json:"created_at"`
Items          []TerminalItem `json:"items" binding:"required"`
}

type TerminalItem struct {
ItemID     string  `json:"item_id"    binding:"required"`
Quantity   int     `json:"quantity"   binding:"required"`
UnitPrice  float64 `json:"unit_price" binding:"required"`
TotalPrice float64 `json:"total_price" binding:"required"`
}

type SyncPushRequest struct {
ShopID       string                `json:"shop_id"       binding:"required"`
Transactions []TerminalTransaction `json:"transactions"  binding:"required"`
}

type SyncPushResult struct {
Accepted  int      `json:"accepted"`
Duplicate int      `json:"duplicate"`
Failed    int      `json:"failed"`
Errors    []string `json:"errors,omitempty"`
}

type CatalogDelta struct {
Items      interface{} `json:"items"`
Categories interface{} `json:"categories"`
AsOf       string      `json:"as_of"`
}

func PushTransactions(req SyncPushRequest) (SyncPushResult, error) {
result := SyncPushResult{}

for _, t := range req.Transactions {
// Check for duplicate
var count int
db.DB.Get(&count,
`SELECT COUNT(*) FROM transactions WHERE terminal_txn_id=$1 AND shop_id=$2`,
t.TerminalTxnID, req.ShopID)
if count > 0 {
result.Duplicate++
continue
}

// Parse created_at or use now
createdAt := time.Now()
if t.CreatedAt != "" {
if parsed, err := time.Parse("2006-01-02T15:04:05", t.CreatedAt); err == nil {
createdAt = parsed
}
}

tx, err := db.DB.Beginx()
if err != nil {
result.Failed++
result.Errors = append(result.Errors, "db tx error: "+err.Error())
continue
}

var txnID string
err = tx.QueryRow(
`INSERT INTO transactions
 (shop_id, staff_id, total_amount, discount_amount, payment_mode,
  payment_status, cash_received, change_given,
  synced_from_terminal, terminal_txn_id, created_at)
 VALUES ($1,$2,$3,$4,$5,'confirmed',$6,$7,true,$8,$9)
 RETURNING id`,
t.ShopID, t.StaffID, t.TotalAmount, t.DiscountAmount,
t.PaymentMode, t.CashReceived, t.ChangeGiven,
t.TerminalTxnID, createdAt,
).Scan(&txnID)
if err != nil {
tx.Rollback()
result.Failed++
result.Errors = append(result.Errors, t.TerminalTxnID+": "+err.Error())
continue
}

// Insert items + deduct stock
for _, item := range t.Items {
tx.Exec(
`INSERT INTO transaction_items
 (transaction_id, item_id, quantity, unit_price, total_price)
 VALUES ($1,$2,$3,$4,$5)`,
txnID, item.ItemID, item.Quantity, item.UnitPrice, item.TotalPrice,
)
tx.Exec(
`UPDATE shop_items SET stock_qty = stock_qty - $1
 WHERE shop_id=$2 AND item_id=$3 AND stock_qty >= $1`,
item.Quantity, req.ShopID, item.ItemID,
)
}

if err := tx.Commit(); err != nil {
result.Failed++
result.Errors = append(result.Errors, t.TerminalTxnID+": commit failed")
continue
}

result.Accepted++
}

// Update terminal sync log
db.DB.Exec(`
INSERT INTO terminal_sync_log (shop_id, last_sync_at, is_online, last_heartbeat)
VALUES ($1, NOW(), true, NOW())
ON CONFLICT (shop_id) DO UPDATE
SET last_sync_at=NOW(), is_online=true, last_heartbeat=NOW()`,
req.ShopID,
)

// Broadcast sync event
go dashboard.PushLiveUpdate("terminal_synced", map[string]interface{}{
"shop_id":  req.ShopID,
"accepted": result.Accepted,
"duplicate": result.Duplicate,
})

return result, nil
}

func PullCatalog(shopID string) (CatalogDelta, error) {
var items []map[string]interface{}
rows, err := db.DB.Queryx(`
SELECT i.id, i.name, i.name_hi, i.price, i.tax_rate, i.category_id, i.is_active
FROM items i
JOIN shop_items si ON si.item_id = i.id
WHERE si.shop_id = $1
ORDER BY i.name`, shopID)
if err != nil {
return CatalogDelta{}, err
}
defer rows.Close()
for rows.Next() {
row := make(map[string]interface{})
rows.MapScan(row)
// convert []byte values to string
for k, v := range row {
if b, ok := v.([]byte); ok {
row[k] = string(b)
}
}
items = append(items, row)
}

var cats []map[string]interface{}
catRows, err := db.DB.Queryx(`SELECT id, name, name_hi FROM categories ORDER BY name`)
if err != nil {
return CatalogDelta{}, err
}
defer catRows.Close()
for catRows.Next() {
row := make(map[string]interface{})
catRows.MapScan(row)
for k, v := range row {
if b, ok := v.([]byte); ok {
row[k] = string(b)
}
}
cats = append(cats, row)
}

return CatalogDelta{
Items:      items,
Categories: cats,
AsOf:       time.Now().Format(time.RFC3339),
}, nil
}

func Heartbeat(shopID string, queueSize int) error {
_, err := db.DB.Exec(`
INSERT INTO terminal_sync_log (shop_id, last_heartbeat, is_online, pending_queue_size)
VALUES ($1, NOW(), true, $2)
ON CONFLICT (shop_id) DO UPDATE
SET last_heartbeat=NOW(), is_online=true, pending_queue_size=$2`,
shopID, queueSize,
)
return err
}

func GetTerminalStatus() ([]map[string]interface{}, error) {
rows, err := db.DB.Queryx(`
SELECT
tsl.shop_id,
s.name                                       AS shop_name,
tsl.last_heartbeat,
tsl.last_sync_at,
tsl.pending_queue_size,
tsl.is_online,
(NOW() - tsl.last_heartbeat) > INTERVAL '2 minutes' AS is_stale
FROM terminal_sync_log tsl
JOIN shops s ON s.id = tsl.shop_id
ORDER BY tsl.last_heartbeat DESC`)
if err != nil {
return nil, err
}
defer rows.Close()
var result []map[string]interface{}
for rows.Next() {
row := make(map[string]interface{})
rows.MapScan(row)
for k, v := range row {
if b, ok := v.([]byte); ok {
row[k] = string(b)
}
}
result = append(result, row)
}
return result, nil
}
