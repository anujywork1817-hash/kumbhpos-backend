package payment

import (
"crypto/hmac"
"crypto/sha256"
"encoding/hex"
"errors"
"fmt"
"os"

"github.com/beellz/kumbhpos/pkg/db"
razorpay "github.com/razorpay/razorpay-go"
)

type QRRequest struct {
ShopID string  `json:"shop_id" binding:"required"`
TxnID  string  `json:"txn_id"  binding:"required"`
Amount float64 `json:"amount"  binding:"required"`
}

type QRResponse struct {
TxnID       string `json:"txn_id"`
OrderID     string `json:"order_id"`
Amount      float64 `json:"amount"`
QRImageURL  string `json:"qr_image_url"`
UPILink     string `json:"upi_link"`
}

type WebhookPayload struct {
Event string `json:"event"`
Payload struct {
Payment struct {
Entity struct {
ID      string `json:"id"`
OrderID string `json:"order_id"`
Amount  int    `json:"amount"`
Status  string `json:"status"`
} `json:"entity"`
} `json:"payment"`
} `json:"payload"`
}

func CreateUPIOrder(req QRRequest) (QRResponse, error) {
client := razorpay.NewClient(
os.Getenv("RAZORPAY_KEY_ID"),
os.Getenv("RAZORPAY_KEY_SECRET"),
)

// Amount in paise (multiply by 100)
amountPaise := int(req.Amount * 100)

orderData := map[string]interface{}{
"amount":          amountPaise,
"currency":        "INR",
"receipt":         req.TxnID,
"payment_capture": 1,
"notes": map[string]interface{}{
"shop_id": req.ShopID,
"txn_id":  req.TxnID,
},
}

order, err := client.Order.Create(orderData, nil)
if err != nil {
return QRResponse{}, fmt.Errorf("razorpay order creation failed: %w", err)
}

orderID := order["id"].(string)
// keyID unused

// Save order ID to transaction
db.DB.Exec(
`UPDATE transactions SET razorpay_order_id=$1 WHERE id=$2`,
orderID, req.TxnID,
)

// UPI payment link — customer can scan or click
upiLink := fmt.Sprintf(
"upi://pay?pa=kumbhpos@razorpay&pn=KumbhPOS&am=%.2f&tr=%s&tn=KumbhPOS%%20Payment",
req.Amount, orderID,
)

// Razorpay hosted QR URL
qrURL := fmt.Sprintf(
"https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=%s",
upiLink,
)

return QRResponse{
TxnID:      req.TxnID,
OrderID:    orderID,
Amount:     req.Amount,
QRImageURL: qrURL,
UPILink:    upiLink,
}, nil
}

func VerifyWebhookSignature(body []byte, signature string) bool {
secret := os.Getenv("RAZORPAY_WEBHOOK_SECRET")
mac := hmac.New(sha256.New, []byte(secret))
mac.Write(body)
expected := hex.EncodeToString(mac.Sum(nil))
return hmac.Equal([]byte(expected), []byte(signature))
}

func ProcessWebhook(payload WebhookPayload) error {
if payload.Event != "payment.captured" {
return nil
}
entity := payload.Payload.Payment.Entity
if entity.Status != "captured" {
return errors.New("payment not captured")
}
_, err := db.DB.Exec(
`UPDATE transactions
 SET payment_status='confirmed', razorpay_payment_id=$1
 WHERE razorpay_order_id=$2`,
entity.ID, entity.OrderID,
)
return err
}

func CheckPaymentStatus(txnID string) (string, error) {
var status string
err := db.DB.Get(&status,
`SELECT payment_status FROM transactions WHERE id=$1`,
txnID)
return status, err
}
