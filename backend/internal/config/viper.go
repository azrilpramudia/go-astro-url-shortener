package config

import (
	"os"

	"github.com/spf13/viper"
)

func NewViper() *viper.Viper {
	v := viper.New()

	v.SetDefault("APP_NAME", "go-astro-url-shortener")
	v.SetDefault("WEB_PORT", 3000)
	v.SetDefault("WEB_PREFORK", false)
	v.SetDefault("LOG_LEVEL", 5)
	v.SetDefault("DB_PORT", 5432)
	v.SetDefault("DB_POOL_IDLE", 10)
	v.SetDefault("DB_POOL_MAX", 100)
	v.SetDefault("DB_POOL_LIFETIME", 300)
	v.SetDefault("WEB_CORS_ORIGIN", "http://localhost:4321")
	v.SetDefault("JWT_TTL_HOURS", 24)
	v.SetDefault("COOKIE_SECURE", false)
	v.SetDefault("RATE_LIMIT_CREATE", 20)
	v.SetDefault("RATE_LIMIT_AUTH", 10)
	v.SetDefault("REDIS_ADDR", "localhost:6379")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("CACHE_TTL_SECONDS", 600)
	v.SetDefault("CLICK_FLUSH_SECONDS", 15)
	v.SetDefault("ANALYTICS_RETENTION_DAYS", 90)
	v.SetDefault("GEOIP_DB_PATH", "")

	v.AutomaticEnv()

	v.SetConfigType("env")
	for _, path := range []string{".env", "../.env"} {
		if _, err := os.Stat(path); err == nil {
			v.SetConfigFile(path)
			if err := v.ReadInConfig(); err != nil {
				panic("fatal error reading .env: " + err.Error())
			}
			break
		}
	}

	return v
}
