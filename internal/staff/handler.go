package staff

import (
"net/http"
"github.com/gin-gonic/gin"
)

func CreateStaffHandler(c *gin.Context) {
var req CreateStaffRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
s, err := CreateStaff(req)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusCreated, s)
}

func ListStaffHandler(c *gin.Context) {
shopID := c.Param("shop_id")
list, err := ListStaffByShop(shopID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, list)
}
