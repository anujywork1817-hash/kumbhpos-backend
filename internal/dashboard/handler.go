package dashboard

import (
"net/http"
"github.com/gin-gonic/gin"
)

func StatsHandler(c *gin.Context) {
stats, err := GetLiveStats()
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, stats)
}

func LeaderboardHandler(c *gin.Context) {
list, err := GetShopLeaderboard()
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, list)
}

func WSHandler(c *gin.Context) {
ServeWS(c.Writer, c.Request)
}
