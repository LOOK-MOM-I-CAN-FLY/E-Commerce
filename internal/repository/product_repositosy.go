package repository

import (
	"sync"

	"github.com/LOOK-MOM-I-CAN-FLY/E-Commerce/internal/models"
)

type ProductRepository struct {
	mu       sync.Mutex
	products map[string]models.Product
}

func NewProductRepository() *ProductRepository {
	repo := &ProductRepository{
		products: make(map[string]models.Product),
	}
	// Инициализация фиктивными данными
	repo.products["1"] = models.Product{
		ID:          "1",
		Name:        "Gin",
		Description: "Фреймворк для языка Go",
		ImageURL:    "web/img/Gin.png",
		Price:       49.99,
	}
	repo.products["2"] = models.Product{
		ID:          "2",
		Name:        "Django",
		Description: "Фреймворк для Python с простым синтаксисом",
		ImageURL:    "web/img/Django.jpg",
		Price:       39.99,
	}
	repo.products["3"] = models.Product{
		ID:          "3",
		Name:        "Spring",
		Description: "Инновационный фреймворк для Java",
		ImageURL:    "web/img/Spring.png",
		Price:       59.99,
	}
	return repo
}

func (r *ProductRepository) GetAll() []models.Product {
	r.mu.Lock()
	defer r.mu.Unlock()
	var products []models.Product
	for _, p := range r.products {
		products = append(products, p)
	}
	return products
}

func (r *ProductRepository) GetByID(id string) (models.Product, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.products[id]
	return p, ok
}
