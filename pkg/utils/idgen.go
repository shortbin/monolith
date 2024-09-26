package utils

import (
	"math/rand"
	"time"
)

func IDGenerator(length int) string {
	//nolint:gosec
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	result := make([]byte, length)

	for i := range result {
		result[i] = chars[r.Intn(len(chars))]
	}

	return string(result)
}
