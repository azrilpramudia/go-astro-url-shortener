package util

import (
	"crypto/rand"
	"math/big"
)

const base62 = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandomCode(n int) (string, error) {
	max := big.NewInt(int64(len(base62)))
	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = base62[idx.Int64()]
	}
	return string(b), nil
}
