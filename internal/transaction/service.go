package transaction

import (
"errors"
"github.com/beellz/kumbhpos/internal/dashboard"
"github.com/beellz/kumbhpos/pkg/db"
)

type CartItem struct {
ItemID   string  `json:"item_id"  binding:"required"`
Quantity int     `json:"quantity" binding:"required"`
}

type CheckoutRequest struct {
ShopID      string     `json:"shop_id"      binding:"required"`
PaymentMode string     `json:"payment_mode" binding:"required"` // upi | cash
Items       []CartItem `json:"items"        binding:"required"`
CashReceived float64   `json:"cash_received"`
DiscountAmount float64 `json:"discount_amount"`
}

type TransactionResponse struct {
ID            string  `db:"id"                json:"id"`
ShopID        string  `db:"shop_id"           json:"shop_id"`
StaffID       string  `db:"staff_id"          json:"staff_id"`
TotalAmount   float64 `db:"total_amount"      json:"total_amount"`
DiscountAmount float64 `db:"discount_amount"  json:"discount_amount"`
PaymentMode   string  `db:"payment_mode"      json:"payment_mode"`
PaymentStatus string  `db:"payment_status"    json:"payment_status"`
CashReceived  float64 `db:"cash_received"     json:"cash_received"`
ChangeGiven   float64 `db:"change_given"      json:"change_given"`
CreatedAt     string  `db:"created_at"        json:"created_at"`
}

type ItemPrice struct {
ID    string  `db:"id"`
Price float64 `db:"price"`
}

func Checkout(req CheckoutRequest, staffID string) (TransactionResponse, error) {
if req.PaymentMode != "upi" && req.PaymentMode != "cash" {
return TransactionResponse{}, errors.New("payment_mode must be upi or cash")
}

// Calculate total from DB prices (never trust client prices)
var total float64
for _, ci := range req.Items {
var item ItemPrice
err := db.DB.Get(&item, `SELECT id, price FROM items WHERE id=$1 AND is_active=true`, ci.ItemID)
if err != nil {
return TransactionResponse{}, errors.New("item not found: " + ci.ItemID)
}
total += item.Price * float64(ci.Quantity)
}
total -= req.DiscountAmount

// Calculate change for cash
var changeGiven float64
if req.PaymentMode == "cash" {
if req.CashReceived < total {
return TransactionResponse{}, errors.New("cash received is less than total amount")
}
changeGiven = req.CashReceived - total
}

// Payment status
status := "pending"
if req.PaymentMode == "cash" {
status = "confirmed"
}

// Begin transaction
tx, err := db.DB.Beginx()
if err != nil {
return TransactionResponse{}, err
}
defer tx.Rollback()

// Insert transaction
var txn TransactionResponse
err = tx.QueryRowx(
`INSERT INTO transactions
 (shop_id, staff_id, total_amount, discount_amount, payment_mode, payment_status, cash_received, change_given)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
 RETURNING id, shop_id, staff_id, total_amount, discount_amount, payment_mode, payment_status, cash_received, change_given, created_at::text`,
req.ShopID, staffID, total, req.DiscountAmount, req.PaymentMode, status,
req.CashReceived, changeGiven,
).StructScan(&txn)
if err != nil {
return TransactionResponse{}, err
}

// Insert transaction items + deduct stock
for _, ci := range req.Items {
var item ItemPrice
db.DB.Get(&item, `SELECT id, price FROM items WHERE id=$1`, ci.ItemID)
lineTotal := item.Price * float64(ci.Quantity)

_, err = tx.Exec(
`INSERT INTO transaction_items (transaction_id, item_id, quantity, unit_price, total_price)
 VALUES ($1,$2,$3,$4,$5)`,
txn.ID, ci.ItemID, ci.Quantity, item.Price, lineTotal,
)
if err != nil {
return TransactionResponse{}, err
}

// Deduct stock
tx.Exec(
`UPDATE shop_items SET stock_qty = stock_qty - $1
 WHERE shop_id=$2 AND item_id=$3 AND stock_qty >= $1`,
ci.Quantity, req.ShopID, ci.ItemID,
)
}

if err = tx.Commit(); err != nil {
return TransactionResponse{}, err
}

// Broadcast live update to dashboard
go dashboard.PushLiveUpdate("new_transaction", map[string]interface{}{
"shop_id":      txn.ShopID,
"amount":       txn.TotalAmount,
"payment_mode": txn.PaymentMode,
"txn_id":       txn.ID,
})
return txn, nil
}

func GetTransaction(txnID string) (TransactionResponse, error) {
var txn TransactionResponse
err := db.DB.Get(&txn,
`SELECT id, shop_id, staff_id, total_amount, discount_amount,
 payment_mode, payment_status, cash_received, change_given, created_at::text
 FROM transactions WHERE id=$1`,
txnID)
return txn, err
}

func ListTransactions(shopID string) ([]TransactionResponse, error) {
var list []TransactionResponse
err := db.DB.Select(&list,
`SELECT id, shop_id, staff_id, total_amount, discount_amount,
 payment_mode, payment_status, cash_received, change_given, created_at::text
 FROM transactions WHERE shop_id=$1 ORDER BY created_at DESC LIMIT 50`,
shopID)
return list, err
}

func ConfirmUPIPayment(txnID, razorpayOrderID, razorpayPaymentID string) error {
_, err := db.DB.Exec(
`UPDATE transactions SET payment_status='confirmed',
 razorpay_order_id=$2, razorpay_payment_id=$3
 WHERE id=$1`,
txnID, razorpayOrderID, razorpayPaymentID,
)
return err
}

type TransactionItem struct {
ID            string  `db:"id"            json:"id"`
TransactionID string  `db:"transaction_id" json:"transaction_id"`
ItemID        string  `db:"item_id"       json:"item_id"`
ItemName      string  `db:"item_name"     json:"item_name"`
Quantity      int     `db:"quantity"      json:"quantity"`
UnitPrice     float64 `db:"unit_price"    json:"unit_price"`
TotalPrice    float64 `db:"total_price"   json:"total_price"`
}

func GetTransactionItems(txnID string) ([]TransactionItem, error) {
var items []TransactionItem
err := db.DB.Select(&items, `
SELECT
ti.id,
ti.transaction_id,
ti.item_id,
i.name  AS item_name,
ti.quantity,
ti.unit_price,
ti.total_price
FROM transaction_items ti
JOIN items i ON i.id = ti.item_id
WHERE ti.transaction_id = $1
ORDER BY i.name`,
txnID)
return items, err
}
