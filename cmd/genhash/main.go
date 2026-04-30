package main

import (
"fmt"
"golang.org/x/crypto/bcrypt"
)

func main() {
hash1 := "$2a$10$fPJRHkcsV4/2xsEbQQgBl.j5C/VowYpBDhZFi./gAzbRzwHXyAqrW"
hash2 := "$2a$10$Cur.aSLsTCUxYeC6WAVEqOsRr1Xj8MxtD/mbqfqyKpy2qrfpYV7Le"

err1 := bcrypt.CompareHashAndPassword([]byte(hash1), []byte("1817"))
err2 := bcrypt.CompareHashAndPassword([]byte(hash2), []byte("1718"))

fmt.Println("PIN 1817 vs hash1:", err1)
fmt.Println("PIN 1718 vs hash2:", err2)
}
