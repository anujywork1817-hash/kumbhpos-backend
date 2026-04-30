package jobs

import (
"net/http"
"github.com/gin-gonic/gin"
)

func ManualEODHandler(c *gin.Context) {
date := c.Query("date")
if err := ManualEOD(date); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, gin.H{"message": "EOD report generated and broadcast"})
}

func SchedulerStatusHandler(c *gin.Context) {
entries := Scheduler.Entries()
var result []map[string]interface{}
for _, e := range entries {
result = append(result, map[string]interface{}{
"id":       e.ID,
"next_run": e.Next.Format("2006-01-02 15:04:05"),
"prev_run": e.Prev.Format("2006-01-02 15:04:05"),
})
}
c.JSON(http.StatusOK, gin.H{
"total_jobs": len(entries),
"jobs":       result,
})
}
