package util

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"strings"
)

const VisitorCookieName = "url_shortener_visitor"

func NewVisitorToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func IsVisitorToken(value string) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil && len(decoded) == 32
}

func HashVisitorToken(secret, token string) string {
	hash := hmac.New(sha256.New, []byte(secret))
	_, _ = hash.Write([]byte(token))
	return hex.EncodeToString(hash.Sum(nil))
}

func ReferrerHost(value string) string {
	referrer, err := url.Parse(value)
	if err != nil {
		return ""
	}
	return strings.ToLower(referrer.Hostname())
}

func DeviceType(userAgent string) string {
	userAgent = strings.ToLower(userAgent)
	switch {
	case userAgent == "":
		return "unknown"
	case strings.Contains(userAgent, "bot"),
		strings.Contains(userAgent, "crawler"),
		strings.Contains(userAgent, "spider"):
		return "bot"
	case strings.Contains(userAgent, "ipad"),
		strings.Contains(userAgent, "tablet"),
		strings.Contains(userAgent, "kindle"):
		return "tablet"
	case strings.Contains(userAgent, "mobile"),
		strings.Contains(userAgent, "iphone"),
		strings.Contains(userAgent, "ipod"),
		strings.Contains(userAgent, "android"):
		return "mobile"
	default:
		return "desktop"
	}
}
