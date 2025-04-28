package controllers

import (
	"digital-marketplace/internal/database"
	"digital-marketplace/internal/models"
	"digital-marketplace/internal/services"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
)

type BuyController struct {
	validationService *services.ValidationService
}

func NewBuyController() *BuyController {
	return &BuyController{
		validationService: services.NewValidationService(),
	}
}

// Функция валидации ID продукта можно удалить, т.к. она перенесена в ValidationService
func validateProductID(idStr string) (uint, error) {
	// Удаляем пробелы и проверяем на пустоту
	idStr = strings.TrimSpace(idStr)
	if idStr == "" {
		return 0, fmt.Errorf("ID продукта не указан")
	}

	// Преобразуем строку в число
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Неверный формат ID продукта")
	}

	// Проверяем на допустимый диапазон
	if id == 0 || id > 1000000 { // Выберите подходящий верхний порог
		return 0, fmt.Errorf("ID продукта вне допустимого диапазона")
	}

	return uint(id), nil
}

// ShowBuyPage displays the confirmation page for buying a specific product
func (bc *BuyController) ShowBuyPage(c *gin.Context) {
	productIDStr := c.Param("productID") // Get product ID from URL parameter

	// Базовая валидация - конвертация строки в uint
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		renderTemplate(c, "error.html", gin.H{"Error": "Некорректный ID продукта"})
		return
	}

	// Дополнительная валидация с использованием ValidationService
	validID, errMsg := bc.validationService.ValidateProductID(uint(productID))
	if !validID {
		renderTemplate(c, "error.html", gin.H{"Error": errMsg})
		return
	}

	var product models.Product
	// Use Preload to fetch user info if needed in the template later
	result := database.DB.First(&product, productID)
	if result.Error != nil {
		renderTemplate(c, "error.html", gin.H{"Error": "Товар не найден"})
		return
	}

	// Проверка, что товар не принадлежит текущему пользователю
	user, userExists := getUserFromContext(c)
	if userExists && user.ID == product.UserID {
		renderTemplate(c, "error.html", gin.H{"Error": "Вы не можете купить свой собственный товар"})
		return
	}

	renderTemplate(c, "buy.html", gin.H{
		"Product": product,
	})
}

// HandleBuy processes the purchase request, creates an order, and sends confirmation
func (bc *BuyController) HandleBuy(c *gin.Context) {
	productIDStr := c.Param("productID") // Get product ID from URL parameter

	// Базовая валидация - конвертация строки в uint
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		renderTemplate(c, "error.html", gin.H{"Error": "Некорректный ID продукта"})
		return
	}

	// Дополнительная валидация с использованием ValidationService
	validID, errMsg := bc.validationService.ValidateProductID(uint(productID))
	if !validID {
		renderTemplate(c, "error.html", gin.H{"Error": errMsg})
		return
	}

	// Get user from context (set by AuthRequired middleware)
	user, exists := getUserFromContext(c)
	if !exists {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var product models.Product
	// Fetch the product again to ensure it exists
	if err := database.DB.First(&product, productID).Error; err != nil {
		renderTemplate(c, "error.html", gin.H{"Error": "Товар не найден"})
		return
	}

	// Проверка, что товар не принадлежит текущему пользователю
	if user.ID == product.UserID {
		renderTemplate(c, "error.html", gin.H{"Error": "Вы не можете купить свой собственный товар"})
		return
	}

	// Проверка, не куплен ли товар уже ранее
	var existingOrder models.OrderItem
	alreadyPurchased := database.DB.
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("order_items.product_id = ? AND orders.user_id = ?", productID, user.ID).
		First(&existingOrder).Error == nil

	if alreadyPurchased {
		renderTemplate(c, "error.html", gin.H{"Error": "Вы уже приобрели этот товар ранее"})
		return
	}

	// --- Create Order and OrderItem in Transaction ---
	tx := database.DB.Begin()

	// 1. Create Order
	order := models.Order{
		UserID:    user.ID,
		CreatedAt: time.Now(),
	}
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		fmt.Printf("Error creating order for product %d by user %d: %v\n", product.ID, user.ID, err)
		renderTemplate(c, "buy.html", gin.H{
			"Product": product,
			"Error":   "Не удалось создать заказ. Попробуйте снова.",
		})
		return
	}

	// 2. Create OrderItem
	orderItem := models.OrderItem{
		OrderID:   order.ID,
		ProductID: product.ID,
	}
	if err := tx.Create(&orderItem).Error; err != nil {
		tx.Rollback()
		fmt.Printf("Error creating order item for order %d, product %d: %v\n", order.ID, product.ID, err)
		renderTemplate(c, "buy.html", gin.H{
			"Product": product,
			"Error":   "Не удалось добавить товар в заказ. Попробуйте снова.",
		})
		return
	}

	// 3. Commit Transaction
	fmt.Println("Attempting to commit transaction for single product purchase...")
	if err := tx.Commit().Error; err != nil {
		fmt.Println("Error committing transaction:", err)
		// Rollback is implicitly done if Commit fails
		renderTemplate(c, "buy.html", gin.H{
			"Product": product,
			"Error":   "Ошибка сохранения заказа. Попробуйте снова.",
		})
		return
	}
	fmt.Println("Transaction committed successfully!")

	// 4. Send confirmation email (using the copied function)
	// Валидация email перед отправкой
	if valid, _ := bc.validationService.ValidateEmail(user.Email); valid {
		go sendOrderConfirmationEmail(user.Email, order.ID)
	} else {
		fmt.Printf("Предупреждение: некорректный email пользователя %d: %s\n", user.ID, user.Email)
	}

	// 5. Redirect to a generic success page
	c.Redirect(http.StatusFound, "/order/success/")
}

// --- Copied Email Sending Logic (Example using gomail) ---
// Note: Ideally, this should be in a shared service package.
// Adjusted to not require filePaths parameter as it fetches items by orderID.
func sendOrderConfirmationEmail(toEmail string, orderID uint) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	fromEmail := os.Getenv("SMTP_FROM_EMAIL") // Email to send from
	baseURL := os.Getenv("BASE_URL")

	if fromEmail == "" {
		fromEmail = "orders@digital-marketplace.com" // Default sender
	}
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Default base URL
	}

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

	body := fmt.Sprintf(`Уважаемый клиент!

Спасибо за ваш заказ #%d в Digital Marketplace!

Ваш заказ успешно обработан. Ниже приведены ссылки для скачивания приобретенных товаров:

`, orderID)

	var orderItems []models.OrderItem
	// Fetch order items (including the product details) for the specific order
	dbResult := database.DB.Preload("Product").Where("order_id = ?", orderID).Find(&orderItems)
	if dbResult.Error != nil {
		fmt.Printf("Ошибка получения товаров для заказа %d при отправке email: %v\n", orderID, dbResult.Error)
		// Decide if you want to send the email without product links or just return
		return
	}

	fileService := services.NewFileService()

	if len(orderItems) > 0 {
		for i, item := range orderItems {
			product := item.Product
			downloadToken, tokenErr := fileService.GenerateDownloadToken(product.ID)
			if tokenErr != nil {
				fmt.Printf("Ошибка создания токена для продукта %d (заказ %d): %v\n", product.ID, orderID, tokenErr)
				continue // Skip this item if token generation fails
			}
			downloadURL := fileService.GenerateDownloadURL(downloadToken, baseURL)
			body += fmt.Sprintf("%d. %s: %s (ссылка действительна 24 часа)\n", i+1, product.Title, downloadURL)
		}
	} else {
		body += "Не удалось найти информацию о товарах в этом заказе.\n"
	}

	body += `
С уважением,
Команда Digital Marketplace`
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	if smtpHost == "mailhog" { // Specific handling for mailhog
		d.SSL = false
	}

	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("Не удалось отправить email подтверждения заказа #%d на %s: %v\n", orderID, toEmail, err)
	} else {
		fmt.Printf("Email подтверждения заказа #%d успешно отправлен на %s\n", orderID, toEmail)
	}
}
