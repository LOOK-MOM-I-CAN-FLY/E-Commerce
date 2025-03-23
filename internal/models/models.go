package models

// User – модель пользователя
type User struct {
	Email    string
	Password string
}

// Product – модель товара
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	Price       float64 `json:"price"`
}

// Order – модель заказа (если потребуется хранение заказов)
type Order struct {
	ID        string
	UserEmail string
	Products  []Product
}
