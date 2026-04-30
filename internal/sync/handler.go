package sync

import (
"net/http"
"github.com/gin-gonic/gin"
)

func PushHandler(c *gin.Context) {
var req SyncPushRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
result, err := PushTransactions(req)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, result)
}

func PullHandler(c *gin.Context) {
shopID := c.Query("shop_id")
if shopID == "" {
shopID = c.GetString("shop_id")
}
delta, err := PullCatalog(shopID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, delta)
}

func HeartbeatHandler(c *gin.Context) {
var body struct {
ShopID    string `json:"shop_id"     binding:"required"`
QueueSize int    `json:"queue_size"`
}
if err := c.ShouldBindJSON(&body); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
if err := Heartbeat(body.ShopID, body.QueueSize); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func TerminalStatusHandler(c *gin.Context) {
status, err := GetTerminalStatus()
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, status)
}
