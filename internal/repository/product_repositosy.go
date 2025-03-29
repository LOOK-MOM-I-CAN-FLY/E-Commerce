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
		Description: "Высокопроизводительный фреймворк для веб-приложений на Go. Обеспечивает отличную производительность и низкий расход памяти.",
		ImageURL:    "https://raw.githubusercontent.com/gin-gonic/logo/master/color.png",
		Price:       49.99,
	}
	repo.products["2"] = models.Product{
		ID:          "2",
		Name:        "Django",
		Description: "Мощный Python-фреймворк для создания веб-приложений. Включает ORM, админ-панель и множество полезных инструментов.",
		ImageURL:    "https://static.djangoproject.com/img/logos/django-logo-positive.png",
		Price:       39.99,
	}
	repo.products["3"] = models.Product{
		ID:          "3",
		Name:        "Spring",
		Description: "Полнофункциональный фреймворк для Java с множеством модулей. Идеально подходит для корпоративных приложений.",
		ImageURL:    "https://spring.io/img/spring-2.svg",
		Price:       59.99,
	}
	repo.products["4"] = models.Product{
		ID:          "4",
		Name:        "React",
		Description: "JavaScript-библиотека для создания пользовательских интерфейсов. Популярное решение для современных веб-приложений.",
		ImageURL:    "https://upload.wikimedia.org/wikipedia/commons/thumb/a/a7/React-icon.svg/1200px-React-icon.svg.png",
		Price:       45.99,
	}
	repo.products["5"] = models.Product{
		ID:          "5",
		Name:        "Vue.js",
		Description: "Прогрессивный JavaScript-фреймворк для создания пользовательских интерфейсов. Легко интегрируется в проекты.",
		ImageURL:    "https://upload.wikimedia.org/wikipedia/commons/thumb/9/95/Vue.js_Logo_2.svg/1200px-Vue.js_Logo_2.svg.png",
		Price:       42.99,
	}
	repo.products["6"] = models.Product{
		ID:          "6",
		Name:        "Laravel",
		Description: "Элегантный PHP-фреймворк для веб-разработки. Предлагает отличную экосистему и множество готовых решений.",
		ImageURL:    "https://upload.wikimedia.org/wikipedia/commons/thumb/9/9a/Laravel.svg/985px-Laravel.svg.png",
		Price:       47.99,
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
