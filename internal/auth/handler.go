package auth

import (
"net/http"

"github.com/gin-gonic/gin"
)

func LoginHandler(c *gin.Context) {
var req LoginRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
result, err := Login(req)
if err != nil {
c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, result)
}
