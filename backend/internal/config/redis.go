package config

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewRedis(v *viper.Viper, log *logrus.Logger) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         v.GetString("REDIS_ADDR"),
		Password:     v.GetString("REDIS_PASSWORD"),
		DB:           v.GetInt("REDIS_DB"),
		DialTimeout:  500 * time.Millisecond,
		ReadTimeout:  300 * time.Millisecond,
		WriteTimeout: 300 * time.Millisecond,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Warnf("redis not reachable, falling back to database only: %v", err)
	}

	return client
}
