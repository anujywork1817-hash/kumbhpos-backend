package middleware

import (
"net/http"
"github.com/gin-gonic/gin"
)

func AdminRequired() gin.HandlerFunc {
return func(c *gin.Context) {
role := c.GetString("role")
if role != "admin" && role != "manager" {
c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin or manager access required"})
return
}
c.Next()
}
}
