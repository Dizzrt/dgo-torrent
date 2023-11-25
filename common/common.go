package common

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GeneratePeerID() string {
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charSetLength := big.NewInt(int64(len(charSet)))

	randomString := make([]byte, 14)
	for i := 0; i < 14; i++ {
		index, err := rand.Int(rand.Reader, charSetLength)
		if err != nil {
			panic(err)
		}
		randomString[i] = charSet[index.Int64()]
	}

	return fmt.Sprintf("DT-%s-%s", "00", string(randomString))
}
