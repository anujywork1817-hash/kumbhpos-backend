package admin

import (
"net/http"
"github.com/gin-gonic/gin"
)

func GlobalStatsHandler(c *gin.Context) {
stats, err := GetGlobalStats()
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, stats)
}

func ListShopsHandler(c *gin.Context) {
shops, err := ListShopsDetailed()
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, shops)
}

func ToggleShopHandler(c *gin.Context) {
shopID := c.Param("id")
var body struct {
Active bool `json:"is_active"`
}
if err := c.ShouldBindJSON(&body); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
if err := SetShopActive(shopID, body.Active); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, gin.H{"message": "shop updated"})
}

func ListStaffHandler(c *gin.Context) {
list, err := ListAllStaff()
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, list)
}

func ResetPINHandler(c *gin.Context) {
staffID := c.Param("id")
var body struct {
NewPIN string `json:"new_pin" binding:"required"`
}
if err := c.ShouldBindJSON(&body); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
if err := ResetStaffPIN(staffID, body.NewPIN); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, gin.H{"message": "PIN reset successful"})
}

func ToggleStaffHandler(c *gin.Context) {
staffID := c.Param("id")
var body struct {
Active bool `json:"is_active"`
}
if err := c.ShouldBindJSON(&body); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
if err := SetStaffActive(staffID, body.Active); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, gin.H{"message": "staff updated"})
}

func ChangeRoleHandler(c *gin.Context) {
staffID := c.Param("id")
var body struct {
Role string `json:"role" binding:"required"`
}
if err := c.ShouldBindJSON(&body); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
if err := ChangeStaffRole(staffID, body.Role); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

func ListItemsHandler(c *gin.Context) {
items, err := ListAllItems()
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, items)
}

func UpdatePriceHandler(c *gin.Context) {
itemID := c.Param("id")
var body struct {
Price float64 `json:"price" binding:"required"`
}
if err := c.ShouldBindJSON(&body); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
if err := UpdateItemPrice(itemID, body.Price); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, gin.H{"message": "price updated"})
}

func ToggleItemHandler(c *gin.Context) {
itemID := c.Param("id")
var body struct {
Active bool `json:"is_active"`
}
if err := c.ShouldBindJSON(&body); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
if err := SetItemActive(itemID, body.Active); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, gin.H{"message": "item updated"})
}
