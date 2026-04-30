package shop

import (
"net/http"
"github.com/gin-gonic/gin"
)

func CreateShopHandler(c *gin.Context) {
var req CreateShopRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
s, err := CreateShop(req)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusCreated, s)
}

func ListShopsHandler(c *gin.Context) {
shops, err := ListShops()
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, shops)
}
