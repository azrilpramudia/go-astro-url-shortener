package entity

import "time"

type User struct {
	ID           uint      `gorm:"column:id;primaryKey;autoIncrement"`
	Name         string    `gorm:"column:name"`
	Email        string    `gorm:"column:email"`
	PasswordHash string    `gorm:"column:password_hash"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
}

func (u *User) TableName() string {
	return "users"
}
