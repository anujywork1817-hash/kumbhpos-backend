package transaction

import (
"net/http"
"github.com/gin-gonic/gin"
)

func CheckoutHandler(c *gin.Context) {
var req CheckoutRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
staffID := c.GetString("staff_id")
txn, err := Checkout(req, staffID)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusCreated, txn)
}

func GetTransactionHandler(c *gin.Context) {
txn, err := GetTransaction(c.Param("id"))
if err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
return
}
c.JSON(http.StatusOK, txn)
}

func ListTransactionsHandler(c *gin.Context) {
shopID := c.Query("shop_id")
if shopID == "" {
shopID = c.GetString("shop_id")
}
list, err := ListTransactions(shopID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, list)
}

func ConfirmUPIHandler(c *gin.Context) {
txnID := c.Param("id")
var body struct {
RazorpayOrderID   string `json:"razorpay_order_id"`
RazorpayPaymentID string `json:"razorpay_payment_id"`
}
if err := c.ShouldBindJSON(&body); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
if err := ConfirmUPIPayment(txnID, body.RazorpayOrderID, body.RazorpayPaymentID); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, gin.H{"message": "payment confirmed"})
}
func GetTransactionItemsHandler(c *gin.Context) {
txnID := c.Param("id")
items, err := GetTransactionItems(txnID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, items)
}
