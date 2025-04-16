package models

import "time"

// Order represents a customer order
type Order struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"not null"` // Foreign key to User
	CreatedAt time.Time
	// Add other fields if needed, e.g., TotalAmount, Status

	User  User        `gorm:"foreignKey:UserID"`
	Items []OrderItem `gorm:"foreignKey:OrderID"` // One-to-many relationship
}
