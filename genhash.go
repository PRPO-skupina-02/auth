package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	passwords := []string{"admin123", "employee123", "customer123"}
	for _, pwd := range passwords {
		hash, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		fmt.Printf("Password: %s\nHash: %s\n\n", pwd, string(hash))
	}
}
