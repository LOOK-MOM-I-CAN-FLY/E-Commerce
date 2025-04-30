package controllers

import (
	"digital-marketplace/internal/database"
	"digital-marketplace/internal/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
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

	// 3. Рассчитываем общую стоимость товаров в корзине
	var totalPrice float64 = 0
	for _, item := range cartItems {
		totalPrice += item.Product.Price
	}

	// 4. Проверяем, достаточно ли средств у пользователя
	if user.Balance < totalPrice {
		// Если средств недостаточно, перенаправляем обратно на страницу корзины с ошибкой
		c.Set("cart_error", "Недостаточно средств на балансе. Пожалуйста, заработайте больше кредитов.")
		c.Redirect(http.StatusFound, "/cart")
		return
	}

	// 5. Create Order and OrderItems within a transaction
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

	// 6. Вычитаем стоимость покупки из баланса пользователя
	user.Balance -= totalPrice
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		fmt.Println("Error updating user balance:", err)
		c.Redirect(http.StatusFound, "/cart")
		return
	}

	// 7. Clear the user's cart
	if err := tx.Where("user_id = ?", user.ID).Delete(&models.CartItem{}).Error; err != nil {
		tx.Rollback()
		fmt.Println("Error clearing cart:", err)
		c.Redirect(http.StatusFound, "/cart") // Add error feedback
		return
	}

	// 8. Commit transaction
	fmt.Println("Attempting to commit transaction for order creation...") // Лог перед коммитом
	if err := tx.Commit().Error; err != nil {
		// Важно логировать ошибку коммита!
		fmt.Println("Error committing transaction:", err)
		// Rollback здесь уже не нужен, так как Commit не удался
		c.Redirect(http.StatusFound, "/cart") // Add error feedback
		return
	}
	fmt.Println("Transaction committed successfully!") // Лог после успешного коммита

	// 9. Send confirmation email (Now calls the function defined in buy_controller.go indirectly via package scope or needs refactoring)
	go sendOrderConfirmationEmail(user.Email, order.ID)

	// 10. Redirect to a success page (or dashboard)
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
