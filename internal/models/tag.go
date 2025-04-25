package models

// Tag represents a tag that can be applied to products.
type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ProductTag represents the many-to-many relationship between products and tags.
// Пока не используется напрямую в JSON ответах, но нужна для связи в БД.
type ProductTag struct {
	ProductID uint `gorm:"primaryKey"`
	TagID     int  `gorm:"primaryKey"`
}
