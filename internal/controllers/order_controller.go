package controllers

import (
	"digital-marketplace/internal/database"
	"digital-marketplace/internal/models"
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
	// Maybe fetch order details later if needed by ID
	renderTemplate(c, "order_success.html", gin.H{})
}

// --- Email Sending Logic (Example using gomail) ---

func sendOrderConfirmationEmail(toEmail string, orderID uint, filePaths []string) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	fromEmail := os.Getenv("SMTP_FROM_EMAIL") // Email to send from

	if smtpHost == "" || smtpPortStr == "" || smtpUser == "" || smtpPass == "" || fromEmail == "" {
		fmt.Println("SMTP environment variables not set. Skipping email.")
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

	body := fmt.Sprintf("Спасибо за ваш заказ #%d!\n\nВы приобрели следующие цифровые товары:\n", orderID)
	// In a real scenario, you might list product names here.

	m.SetBody("text/plain", body)

	// Attach files
	for _, path := range filePaths {
		// IMPORTANT: Assume 'path' is relative to the 'uploads' dir if stored like '/uploads/file.ext'
		// Construct the full system path.
		fullPath := "." + path // Assuming relative path from project root e.g., ./uploads/file.ext
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			fmt.Printf("Email attachment file not found: %s. Skipping attachment.\n", fullPath)
			continue // Skip attaching this file
		}
		m.Attach(fullPath)
	}

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		fmt.Println("Failed to send order confirmation email:", err)
	} else {
		fmt.Println("Order confirmation email sent successfully to", toEmail)
	}
}
