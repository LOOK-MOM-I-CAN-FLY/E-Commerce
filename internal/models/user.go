package models

import "time"

type User struct {
	ID        uint    `gorm:"primaryKey"`
	Username  string  `gorm:"size:255"`
	Email     string  `gorm:"unique;not null"`
	Password  string  `gorm:"not null"`
	Balance   float64 `gorm:"default:0"`
	CreatedAt time.Time
}
