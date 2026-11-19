package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVisitorTokenIsRandomAndHashed(t *testing.T) {
	first, err := NewVisitorToken()
	require.NoError(t, err)
	second, err := NewVisitorToken()
	require.NoError(t, err)
	require.NotEqual(t, first, second)
	require.True(t, IsVisitorToken(first))
	require.False(t, IsVisitorToken("arbitrary"))

	hash := HashVisitorToken(strings.Repeat("s", 32), first)
	require.Len(t, hash, 64)
	require.NotContains(t, hash, first)
	require.Equal(t, hash, HashVisitorToken(strings.Repeat("s", 32), first))
	require.NotEqual(t, hash, HashVisitorToken(strings.Repeat("t", 32), first))
}

func TestReferrerHostStoresOnlyHost(t *testing.T) {
	require.Equal(t, "news.example", ReferrerHost("https://News.Example/story?id=private"))
	require.Empty(t, ReferrerHost(""))
	require.Empty(t, ReferrerHost("not a url"))
}

func TestDeviceType(t *testing.T) {
	tests := map[string]string{
		"":                                "unknown",
		"Googlebot/2.1":                   "bot",
		"Mozilla/5.0 (iPad)":              "tablet",
		"Mozilla/5.0 (iPhone) Mobile":     "mobile",
		"Mozilla/5.0 (X11; Linux x86_64)": "desktop",
	}

	for userAgent, expected := range tests {
		t.Run(expected, func(t *testing.T) {
			require.Equal(t, expected, DeviceType(userAgent))
		})
	}
}
