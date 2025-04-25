package models

import "time"

type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	Price       float64   `gorm:"not null;default:0" json:"price"`
	FilePath    string    `gorm:"not null" json:"filePath"`
	ImagePath   string    `json:"imagePath"`
	UserID      uint      `json:"-"`
	CreatedAt   time.Time `json:"createdAt"`
}
