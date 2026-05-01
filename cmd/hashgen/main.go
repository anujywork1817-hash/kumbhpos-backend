package main

import (
"fmt"
"golang.org/x/crypto/bcrypt"
)

func main() {
pins := []string{"1817", "1718", "1817145"}
for _, pin := range pins {
hash, _ := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
fmt.Printf("PIN %s: %s\n", pin, string(hash))
}
}
