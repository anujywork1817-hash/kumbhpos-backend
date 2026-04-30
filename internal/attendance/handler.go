package attendance

import (
"net/http"
"github.com/gin-gonic/gin"
)

func ClockInHandler(c *gin.Context) {
var req ClockInRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
rec, err := ClockIn(req)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusCreated, rec)
}

func ClockOutHandler(c *gin.Context) {
var req ClockOutRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
rec, err := ClockOut(req)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, rec)
}

func ListAttendanceHandler(c *gin.Context) {
shopID := c.Query("shop_id")
date := c.Query("date")
list, err := GetAttendance(shopID, date)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
if list == nil { list = []Record{} }
c.JSON(http.StatusOK, list)
}

func ActiveShiftsHandler(c *gin.Context) {
shopID := c.Query("shop_id")
list, err := GetActiveShifts(shopID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
if list == nil { list = []Record{} }
c.JSON(http.StatusOK, list)
}

func StaffStatusHandler(c *gin.Context) {
staffID := c.Param("staff_id")
rec, err := GetStaffStatus(staffID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
if rec == nil {
c.JSON(http.StatusOK, gin.H{"clocked_in": false})
return
}
c.JSON(http.StatusOK, gin.H{"clocked_in": true, "record": rec})
}
