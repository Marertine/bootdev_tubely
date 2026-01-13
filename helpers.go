package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

func helperReturn32RandomChars() (string, error) {

	// Create a 32-byte slice
	sliceByte := make([]byte, 32)

	// Fill it with cryptographically secure random data
	if _, err := rand.Read(sliceByte); err != nil {
		return "", errors.New("Unable to create random characters")
	}

	// Encode using base64 URL encoding without padding
	randomString := base64.RawURLEncoding.EncodeToString(sliceByte)

	return randomString, nil
}
