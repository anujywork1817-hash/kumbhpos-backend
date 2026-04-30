package inventory

import (
"net/http"
"github.com/gin-gonic/gin"
)

func StockLevelsHandler(c *gin.Context) {
shopID := c.Query("shop_id")
levels, err := GetStockLevels(shopID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
if levels == nil { levels = []StockLevel{} }
c.JSON(http.StatusOK, levels)
}

func LowStockHandler(c *gin.Context) {
alerts, err := GetLowStockAlerts()
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
if alerts == nil { alerts = []StockLevel{} }
c.JSON(http.StatusOK, alerts)
}

func AdjustStockHandler(c *gin.Context) {
var req AdjustStockRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
if err := AdjustStock(req); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, gin.H{"message": "stock adjusted"})
}

func CreateRestockHandler(c *gin.Context) {
var req CreateRestockRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
if err := SubmitRestockRequest(req); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusCreated, gin.H{"message": "restock request created"})
}

func RestockListHandler(c *gin.Context) {
status := c.Query("status")
list, err := GetRestockRequests(status)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
if list == nil { list = []RestockRequest{} }
c.JSON(http.StatusOK, list)
}

func ApproveRestockHandler(c *gin.Context) {
id := c.Param("id")
if err := ApproveRestockRequest(id); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, gin.H{"message": "restock approved"})
}

func TransferStockHandler(c *gin.Context) {
var body struct {
FromShopID string `json:"from_shop_id" binding:"required"`
ToShopID   string `json:"to_shop_id"   binding:"required"`
ItemID     string `json:"item_id"      binding:"required"`
Qty        int    `json:"qty"          binding:"required"`
}
if err := c.ShouldBindJSON(&body); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
if err := TransferStock(body.FromShopID, body.ToShopID, body.ItemID, body.Qty); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, gin.H{"message": "stock transferred"})
}
