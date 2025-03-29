package service

import (
	"fmt"
	"log"

	"github.com/LOOK-MOM-I-CAN-FLY/E-Commerce/internal/models"
	"github.com/LOOK-MOM-I-CAN-FLY/E-Commerce/internal/repository"
)

type OrderService struct {
	productRepo *repository.ProductRepository
	orderRepo   *repository.OrderRepository
}

func NewOrderService(productRepo *repository.ProductRepository, orderRepo *repository.OrderRepository) *OrderService {
	return &OrderService{
		productRepo: productRepo,
		orderRepo:   orderRepo,
	}
}

func (os *OrderService) Checkout(userEmail, customerEmail string, cart []string) error {
	var purchasedProducts []models.Product
	for _, prodID := range cart {
		if prod, ok := os.productRepo.GetByID(prodID); ok {
			purchasedProducts = append(purchasedProducts, prod)
		}
	}
	if len(purchasedProducts) == 0 {
		return fmt.Errorf("Корзина пуста")
	}

	// Создаём заказ (можно сохранить в репозитории)
	order := models.Order{
		UserEmail: userEmail,
		Products:  purchasedProducts,
	}
	os.orderRepo.Create(order)

	// Имитируем отправку email с фреймворками
	log.Printf("==== ОТПРАВКА ЗАКАЗА ====")
	log.Printf("Отправка фреймворков на email: %s", customerEmail)

	for _, product := range purchasedProducts {
		log.Printf("- %s (%.2f руб)", product.Name, product.Price)
		log.Printf("  Ссылка на изображение: %s", product.ImageURL)
	}

	log.Printf("Общая сумма заказа: %.2f руб", calculateTotal(purchasedProducts))
	log.Printf("==== ЗАКАЗ УСПЕШНО ОФОРМЛЕН ====")

	return nil
}

// Вычисляет общую стоимость заказа
func calculateTotal(products []models.Product) float64 {
	var total float64
	for _, p := range products {
		total += p.Price
	}
	return total
}

// GetProductByID возвращает продукт по его ID
func (s *OrderService) GetProductByID(id string) (models.Product, bool) {
	return s.productRepo.GetByID(id)
}
