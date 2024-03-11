package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func main() {
	b := make([]byte, 16) // 16 bytes = 32 nibbles = 128 bits
	_, err := rand.Read(b)
	var key string
	if err == nil {
		key = hex.EncodeToString(b)
	}
	fmt.Println("token: " + key)
}
