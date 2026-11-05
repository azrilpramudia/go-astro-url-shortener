package config

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func NewDatabase(v *viper.Viper, log *logrus.Logger) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Jakarta",
		v.GetString("DB_HOST"),
		v.GetString("DB_USER"),
		v.GetString("DB_PASSWORD"),
		v.GetString("DB_NAME"),
		v.GetInt("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	conn, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}
	conn.SetMaxIdleConns(v.GetInt("DB_POOL_IDLE"))
	conn.SetMaxOpenConns(v.GetInt("DB_POOL_MAX"))
	conn.SetConnMaxLifetime(time.Second * time.Duration(v.GetInt("DB_POOL_LIFETIME")))

	return db
}
