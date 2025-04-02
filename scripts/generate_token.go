package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// Create the claims
	claims := jwt.MapClaims{
		"sub":  "test-subject-id", // Subject ID matching the test user
		"name": "Test User",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(24 * time.Hour).Unix(), // Token expires in 24 hours
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key from config.yaml
	secretKey := "your-secret-key-change-me-in-production"
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		fmt.Printf("Error signing token: %v\n", err)
		return
	}

	fmt.Println("Generated JWT Token:")
	fmt.Println(signedToken)
}
