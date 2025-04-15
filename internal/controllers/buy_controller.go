package controllers

import (
	"digital-marketplace/internal/database"
	"digital-marketplace/internal/models"
	"digital-marketplace/internal/services"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BuyController struct{}

func NewBuyController() *BuyController {
	return &BuyController{}
}

// ShowBuyPage displays the confirmation page for buying a specific product
func (bc *BuyController) ShowBuyPage(c *gin.Context) {
	productID := c.Param("productID") // Get product ID from URL parameter

	var product models.Product
	result := database.DB.First(&product, productID)
	if result.Error != nil {
		// Handle product not found - maybe render a specific error page or redirect
		renderTemplate(c, "error.html", gin.H{"Error": "Товар не найден"})
		return
	}

	// Pass the product details to the buy.html template
	renderTemplate(c, "buy.html", gin.H{
		"Product": product,
	})
}

// HandleBuy processes the purchase request for a specific product
func (bc *BuyController) HandleBuy(c *gin.Context) {
	productID := c.Param("productID") // Get product ID from URL parameter

	// Get user from context (set by AuthRequired middleware)
	user, exists := getUserFromContext(c)
	if !exists {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var product models.Product
	result := database.DB.First(&product, productID)
	if result.Error != nil {
		c.String(http.StatusNotFound, "Товар не найден") // Keep simple error for now
		return
	}

	// Use the logged-in user's email
	email := user.Email

	err := services.SendProductToEmail(email, product)
	if err != nil {
		// Render error on the buy page
		renderTemplate(c, "buy.html", gin.H{
			"Product": product, // Pass product back to template
			"Error":   fmt.Sprintf("Ошибка отправки файла на email: %v", err),
		})
		return
	}

	// Redirect to a success page with details about the purchase
	c.Redirect(http.StatusFound, "/order/success/?product="+product.Title+"&email="+email)
}
