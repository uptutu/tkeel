package envutil

import (
	"crypto/rand"
	"math/big"
	"os"
)

func Get(key string, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func RandomStr() string {
	bytes := make([]byte, 16)
	for i := 0; i < 16; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(26))
		if err != nil {
			n = big.NewInt(10)
		}
		b := n.Uint64() + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}
