package main

import (
"fmt"
"log"

"github.com/beellz/kumbhpos/pkg/db"
"github.com/beellz/kumbhpos/config"
"golang.org/x/crypto/bcrypt"
)

func main() {
config.Load()
if err := db.Connect(); err != nil {
log.Fatal(err)
}

pins := map[string]string{
"00000000-0000-0000-0000-009699488429": "1817",
"00000000-0000-0000-0000-001234512345": "1718",
}

for id, pin := range pins {
hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
if err != nil {
log.Fatal(err)
}
result, err := db.DB.Exec(
`UPDATE staff SET pin_hash = $1 WHERE id = $2`,
string(hash), id,
)
if err != nil {
log.Fatal(err)
}
rows, _ := result.RowsAffected()
fmt.Printf("Updated %d row for id %s (PIN: %s)\n", rows, id, pin)
}
fmt.Println("Done! PINs fixed.")
}
