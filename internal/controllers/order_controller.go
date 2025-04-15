package controllers

import (
	"digital-marketplace/internal/database"
	"digital-marketplace/internal/models"
	"digital-marketplace/internal/services"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
)

type OrderController struct{}

func NewOrderController() *OrderController {
	return &OrderController{}
}

// Checkout processes the user's cart and creates an order
func (oc *OrderController) Checkout(c *gin.Context) {
	// 1. Get user
	user, exists := getUserFromContext(c)
	if !exists {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// 2. Get cart items for the user (including product data)
	var cartItems []models.CartItem
	cartResult := database.DB.Preload("Product").Where("user_id = ?", user.ID).Find(&cartItems)
	if cartResult.Error != nil || len(cartItems) == 0 {
		fmt.Println("Error fetching cart items or cart empty:", cartResult.Error)
		// Redirect back to cart with an error message (implement flash messages later)
		c.Redirect(http.StatusFound, "/cart")
		return
	}

	// 3. Create Order and OrderItems within a transaction
	tx := database.DB.Begin() // Start transaction

	order := models.Order{
		UserID: user.ID,
	}
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback() // Rollback on error
		fmt.Println("Error creating order:", err)
		c.Redirect(http.StatusFound, "/cart") // Add error feedback
		return
	}

	var orderItems []models.OrderItem
	productFilePaths := []string{} // Slice to hold file paths for email
	for _, item := range cartItems {
		orderItem := models.OrderItem{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			// Optional: Copy Product.Title, Product.FilePath here if needed
		}
		orderItems = append(orderItems, orderItem)
		productFilePaths = append(productFilePaths, item.Product.FilePath) // Collect file paths
	}

	if err := tx.Create(&orderItems).Error; err != nil {
		tx.Rollback()
		fmt.Println("Error creating order items:", err)
		c.Redirect(http.StatusFound, "/cart") // Add error feedback
		return
	}

	// 4. Clear the user's cart
	if err := tx.Where("user_id = ?", user.ID).Delete(&models.CartItem{}).Error; err != nil {
		tx.Rollback()
		fmt.Println("Error clearing cart:", err)
		c.Redirect(http.StatusFound, "/cart") // Add error feedback
		return
	}

	// 5. Commit transaction
	if err := tx.Commit().Error; err != nil {
		fmt.Println("Error committing transaction:", err)
		c.Redirect(http.StatusFound, "/cart") // Add error feedback
		return
	}

	// 6. Send confirmation email with product files/links (Implement this)
	go sendOrderConfirmationEmail(user.Email, order.ID, productFilePaths) // Run in goroutine

	// 7. Redirect to a success page (or dashboard)
	// TODO: Create an order confirmation page
	c.Redirect(http.StatusFound, "/order/success/") // Redirect to a success page
}

// ShowOrderSuccess displays a generic order success page
func (oc *OrderController) ShowOrderSuccess(c *gin.Context) {
	// Получаем информацию о товаре и email из параметров URL
	productTitle := c.Query("product")
	email := c.Query("email")

	data := gin.H{}

	// Если у нас есть информация о товаре, добавляем ее в контекст
	if productTitle != "" {
		data["ProductTitle"] = productTitle
	}

	// Если у нас есть email, добавляем его в контекст
	if email != "" {
		data["Email"] = email
	}

	renderTemplate(c, "order_success.html", data)
}

// --- Email Sending Logic (Example using gomail) ---

func sendOrderConfirmationEmail(toEmail string, orderID uint, filePaths []string) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	fromEmail := os.Getenv("SMTP_FROM_EMAIL") // Email to send from
	baseURL := os.Getenv("BASE_URL")

	if fromEmail == "" {
		fromEmail = "orders@digital-marketplace.com"
	}

	// Если BASE_URL не установлен, используем localhost по умолчанию
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Проверяем, что хост и порт установлены
	if smtpHost == "" || smtpPortStr == "" {
		fmt.Println("SMTP_HOST или SMTP_PORT не установлены. Пропускаем отправку email.")
		return
	}

	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		fmt.Println("Invalid SMTP_PORT:", err)
		return
	}

	m := gomail.NewMessage()
	m.SetHeader("From", fromEmail)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", fmt.Sprintf("Ваш заказ #%d в Digital Marketplace", orderID))

	// Создаем содержимое письма с информацией о заказе и добавляем ссылки на скачивание
	body := fmt.Sprintf(`Уважаемый клиент!

Спасибо за ваш заказ #%d в Digital Marketplace!

Ваш заказ успешно обработан. Ниже приведены ссылки для скачивания приобретенных товаров:

`, orderID)

	// Получаем продукты, связанные с этим заказом
	var orderItems []models.OrderItem
	database.DB.Preload("Product").Where("order_id = ?", orderID).Find(&orderItems)

	// Создаем сервис для генерации безопасных URL
	fileService := services.NewFileService()

	// Если у нас есть товары в заказе, добавляем их в тело письма
	if len(orderItems) > 0 {
		for i, item := range orderItems {
			product := item.Product

			// Создаем защищенный токен для скачивания
			downloadToken, tokenErr := fileService.GenerateDownloadToken(product.ID)
			if tokenErr != nil {
				fmt.Printf("Ошибка создания токена для продукта %d: %v\n", product.ID, tokenErr)
				continue
			}

			// Формируем безопасную ссылку на скачивание с использованием токена
			downloadURL := fileService.GenerateDownloadURL(downloadToken, baseURL)

			body += fmt.Sprintf("%d. %s: %s (ссылка действительна 24 часа)\n",
				i+1, product.Title, downloadURL)
		}
	}

	body += `
С уважением,
Команда Digital Marketplace`

	m.SetBody("text/plain", body)

	// Больше не прикрепляем файлы - используем только ссылки для скачивания
	// Это решает проблему с размером вложений и доступностью

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)

	// Отправляем email
	if err := d.DialAndSend(m); err != nil {
		fmt.Println("Не удалось отправить email подтверждения заказа:", err)
	} else {
		fmt.Println("Email подтверждения заказа успешно отправлен на", toEmail)
	}
}
