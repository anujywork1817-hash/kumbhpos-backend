package payment

import (
"io"
"net/http"

"github.com/gin-gonic/gin"
)

func CreateQRHandler(c *gin.Context) {
var req QRRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
qr, err := CreateUPIOrder(req)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, qr)
}

func WebhookHandler(c *gin.Context) {
body, err := io.ReadAll(c.Request.Body)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
return
}
sig := c.GetHeader("X-Razorpay-Signature")
if !VerifyWebhookSignature(body, sig) {
c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
return
}
var payload WebhookPayload
if err := c.ShouldBindJSON(&payload); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
if err := ProcessWebhook(payload); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func CheckStatusHandler(c *gin.Context) {
txnID := c.Param("id")
status, err := CheckPaymentStatus(txnID)
if err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
return
}
c.JSON(http.StatusOK, gin.H{"txn_id": txnID, "payment_status": status})
}
