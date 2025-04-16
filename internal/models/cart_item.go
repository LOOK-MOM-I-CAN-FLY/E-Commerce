package models

import "time"

// CartItem represents an item in a user's shopping cart
type CartItem struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"not null"` // Foreign key to User
	ProductID uint `gorm:"not null"` // Foreign key to Product
	CreatedAt time.Time

	// Associations (optional but good for GORM)
	User    User    `gorm:"foreignKey:UserID"`
	Product Product `gorm:"foreignKey:ProductID"`
}
