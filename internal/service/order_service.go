package service

import (
	"fmt"
	"log"
	"net/smtp"

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

	// Отправляем email с подтверждением (используем MailHog)
	firstProduct := purchasedProducts[0]
	subject := "Подтверждение покупки"
	body := fmt.Sprintf("Спасибо за покупку!\n\nВы приобрели: %s\n\nФото товара: %s", firstProduct.Name, firstProduct.ImageURL)
	message := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)
	smtpServer := "localhost:1025"
	err := smtp.SendMail(smtpServer, nil, "noreply@example.com", []string{customerEmail}, []byte(message))
	if err != nil {
		log.Printf("Ошибка отправки email: %v", err)
		return err
	}
	log.Printf("Письмо отправлено на %s", customerEmail)
	return nil
}
