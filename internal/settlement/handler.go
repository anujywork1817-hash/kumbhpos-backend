package settlement

import (
"net/http"
"github.com/gin-gonic/gin"
)

func EODReportHandler(c *gin.Context) {
date := c.Query("date")
report, err := GenerateEODReport(date)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, report)
}

func EODPDFHandler(c *gin.Context) {
date := c.Query("date")
report, err := GenerateEODReport(date)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
filename, err := GeneratePDF(report)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.Header("Content-Disposition", "attachment; filename="+filename)
c.File(filename)
}
