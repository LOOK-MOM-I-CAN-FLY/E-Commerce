package models

import "time"

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"size:255"`
	Email     string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
	CreatedAt time.Time
}
