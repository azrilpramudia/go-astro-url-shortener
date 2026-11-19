package util

import (
	"crypto/rand"
	"math/big"
	"strings"
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

func IsValidCode(code string) bool {
	if len(code) == 0 || len(code) > 16 {
		return false
	}
	for i := 0; i < len(code); i++ {
		if !strings.ContainsRune(base62, rune(code[i])) {
			return false
		}
	}
	return true
}

func IsValidAlias(alias string) bool {
	if len(alias) < 3 || len(alias) > 16 {
		return false
	}
	for i := 0; i < len(alias); i++ {
		char := alias[i]
		isAlphaNumeric := char >= 'a' && char <= 'z' ||
			char >= 'A' && char <= 'Z' ||
			char >= '0' && char <= '9'
		if !isAlphaNumeric && char != '-' && char != '_' {
			return false
		}
		if (i == 0 || i == len(alias)-1) && !isAlphaNumeric {
			return false
		}
	}
	return true
}

func IsValidShortCode(code string) bool {
	return IsValidCode(code) || IsValidAlias(code)
}
