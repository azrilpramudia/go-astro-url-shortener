package util

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenManager struct {
	secret []byte
	TTL    time.Duration
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), TTL: ttl}
}

func (m *TokenManager) Generate(userID uint) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(uint64(userID), 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.TTL)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *TokenManager) Parse(tokenString string) (uint, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString, claims,
		func(*jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{"HS256"}), // tolak algoritma lain (mis. "none")
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return 0, errors.New("invalid token subject")
	}
	return uint(id), nil
}
