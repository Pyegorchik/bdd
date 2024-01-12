package random

import (
	"math/rand"
	"time"
)

// Function to generate a random string of a random length
func RandomString() string {
	rand.Seed(time.Now().UnixNano())

	// Define the characters that can be used in the random string
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Generate a random length between 5 and 15 characters
	length := rand.Intn(11) + 5

	// Create a byte slice to store the random string
	randomBytes := make([]byte, length)

	// Fill the byte slice with random characters
	for i := range randomBytes {
		randomBytes[i] = charset[rand.Intn(len(charset))]
	}

	// Convert the byte slice to a string and return it
	return string(randomBytes)
}