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
	pdfBytes, err := GeneratePDFBytes(report)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=settlement-"+date+".pdf")
	c.Header("Content-Type", "application/pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
