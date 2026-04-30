package auth

import (
"errors"
"os"
"time"

"github.com/beellz/kumbhpos/pkg/db"
"github.com/golang-jwt/jwt/v5"
"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
ShopID string `json:"shop_id" binding:"required"`
PIN    string `json:"pin"     binding:"required"`
}

type Staff struct {
ID      string `db:"id"`
ShopID  string `db:"shop_id"`
Name    string `db:"name"`
PinHash string `db:"pin_hash"`
Role    string `db:"role"`
}

func Login(req LoginRequest) (map[string]string, error) {
// Load ALL active staff for this shop
var staffList []Staff
err := db.DB.Select(&staffList,
`SELECT id, shop_id, name, pin_hash, role
 FROM staff WHERE shop_id=$1 AND is_active=true`,
req.ShopID)
if err != nil || len(staffList) == 0 {
return nil, errors.New("staff not found")
}

// Find the one whose PIN matches
var matched *Staff
for i := range staffList {
if bcrypt.CompareHashAndPassword(
[]byte(staffList[i].PinHash), []byte(req.PIN)) == nil {
matched = &staffList[i]
break
}
}
if matched == nil {
return nil, errors.New("invalid PIN")
}

token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
"staff_id": matched.ID,
"shop_id":  matched.ShopID,
"role":     matched.Role,
"exp":      time.Now().Add(12 * time.Hour).Unix(),
})
signed, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
if err != nil {
return nil, err
}

return map[string]string{
"token":      signed,
"staff_id":   matched.ID,
"staff_name": matched.Name,
"role":       matched.Role,
"shop_id":    matched.ShopID,
}, nil
}
