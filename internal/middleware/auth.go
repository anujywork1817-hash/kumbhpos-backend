package middleware

import (
"net/http"
"os"
"strings"

"github.com/gin-gonic/gin"
"github.com/golang-jwt/jwt/v5"
)

func AuthRequired() gin.HandlerFunc {
return func(c *gin.Context) {
// Check header first, then query param (for browser PDF downloads)
tokenStr := ""
header := c.GetHeader("Authorization")
if strings.HasPrefix(header, "Bearer ") {
tokenStr = strings.TrimPrefix(header, "Bearer ")
} else if t := c.Query("token"); t != "" {
tokenStr = t
}

if tokenStr == "" {
c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
return
}

token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
return []byte(os.Getenv("JWT_SECRET")), nil
})
if err != nil || !token.Valid {
c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
return
}
claims := token.Claims.(jwt.MapClaims)
c.Set("staff_id", claims["staff_id"])
c.Set("shop_id", claims["shop_id"])
c.Set("role", claims["role"])
c.Next()
}
}
