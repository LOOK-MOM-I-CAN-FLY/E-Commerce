package controllers

import (
	"digital-marketplace/internal/database"
	"digital-marketplace/internal/models"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CartController struct{}

func NewCartController() *CartController {
	return &CartController{}
}

// AddToCart adds a product to the user's cart
func (cc *CartController) AddToCart(c *gin.Context) {
	// 1. Get user from context (set by AuthRequired middleware)
	user, exists := getUserFromContext(c)
	if !exists {
		// This should technically not happen if AuthRequired is working
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// 2. Get Product ID from URL parameter
	productIDStr := c.Param("productID")
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	// 3. Check if product exists (optional but good practice)
	var product models.Product
	if database.DB.First(&product, uint(productID)).Error != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// 4. Check if item is already in cart (optional: prevent duplicates or increase quantity later)
	var existingItems []models.CartItem
	result := database.DB.Where("user_id = ? AND product_id = ?", user.ID, uint(productID)).Find(&existingItems)

	isInCart := result.RowsAffected > 0

	if isInCart {
		// For now, just redirect back or show a message. Later, could increase quantity.
		fmt.Println("Item already in cart") // Log for now
		// Optionally add a flash message here
		c.Redirect(http.StatusFound, c.Request.Referer()) // Redirect back to previous page
		return
	}

	// 5. Create new cart item
	cartItem := models.CartItem{
		UserID:    user.ID,
		ProductID: uint(productID),
	}

	result = database.DB.Create(&cartItem)
	if result.Error != nil {
		// Handle DB error
		fmt.Println("Error adding item to cart:", result.Error) // Log error
		// Optionally add a flash message for the user
		c.Redirect(http.StatusFound, c.Request.Referer()) // Redirect back
		return
	}

	// 6. Redirect (e.g., back to product page or to the cart)
	fmt.Println("Item added to cart:", cartItem.ID)
	// Optionally add a success flash message
	c.Redirect(http.StatusFound, "/cart") // Redirect to cart page after adding
}

// ShowCart displays the user's shopping cart
func (cc *CartController) ShowCart(c *gin.Context) {
	// 1. Get user from context
	user, exists := getUserFromContext(c)
	if !exists {
		c.Redirect(http.StatusFound, "/login") // Should not happen with AuthRequired
		return
	}

	// 2. Fetch cart items for the user, preloading associated Product data
	var cartItems []models.CartItem
	result := database.DB.Preload("Product").Where("user_id = ?", user.ID).Order("created_at desc").Find(&cartItems)

	if result.Error != nil {
		// Log the error and potentially show an error page or message
		fmt.Println("Error fetching cart items:", result.Error)
		// For now, render the cart page with an error message or empty list
		renderTemplate(c, "cart.html", gin.H{
			"Items": []models.CartItem{}, // Pass empty slice on error
			"Error": "Не удалось загрузить корзину. Попробуйте снова.",
		})
		return
	}

	// 3. Render the cart template with the fetched items
	renderTemplate(c, "cart.html", gin.H{
		"Items": cartItems,
	})
}

// RemoveFromCart removes an item from the user's cart
func (cc *CartController) RemoveFromCart(c *gin.Context) {
	// 1. Get user from context
	user, exists := getUserFromContext(c)
	if !exists {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// 2. Get Cart Item ID from URL parameter
	itemIDStr := c.Param("itemID")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	// 3. Find the cart item to ensure it belongs to the current user
	var cartItem models.CartItem
	result := database.DB.Where("id = ? AND user_id = ?", uint(itemID), user.ID).First(&cartItem)
	if result.Error != nil {
		// Item not found or doesn't belong to the user
		// Log error or redirect with message
		fmt.Println("Error finding cart item to remove:", result.Error)
		c.Redirect(http.StatusFound, "/cart") // Redirect back to cart
		return
	}

	// 4. Delete the item from DB
	deleteResult := database.DB.Delete(&cartItem)
	if deleteResult.Error != nil {
		// Handle DB error during deletion
		fmt.Println("Error removing item from cart:", deleteResult.Error)
		// Optionally add a flash message for the user
		c.Redirect(http.StatusFound, "/cart") // Redirect back
		return
	}

	// 5. Redirect back to cart
	// Optionally add a success flash message
	c.Redirect(http.StatusFound, "/cart")
}
