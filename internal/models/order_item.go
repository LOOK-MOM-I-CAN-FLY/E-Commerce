package models

// OrderItem represents a single item within an Order
type OrderItem struct {
	ID        uint `gorm:"primaryKey"`
	OrderID   uint `gorm:"not null"` // Foreign key to Order
	ProductID uint `gorm:"not null"` // Foreign key to Product
	// Store product details at the time of order (optional but recommended)
	// Example: Title string, FilePath string

	Product Product `gorm:"foreignKey:ProductID"`
}
