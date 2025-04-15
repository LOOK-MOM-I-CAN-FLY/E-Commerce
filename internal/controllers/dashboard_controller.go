package controllers

import (
	"digital-marketplace/internal/database"
	"digital-marketplace/internal/models"

	"github.com/gin-gonic/gin"
)

type DashboardController struct{}

func NewDashboardController() *DashboardController {
	return &DashboardController{}
}

func (dc *DashboardController) ShowDashboard(c *gin.Context) {
	// AuthRequired middleware should have already run and set the user
	user, exists := getUserFromContext(c)
	if !exists {
		// This case should ideally be handled by the middleware redirecting
		// but as a fallback, we can render an error or redirect again.
		// For now, we assume middleware handles the redirect.
		return
	}

	var products []models.Product
	// Fetch products associated with the logged-in user ID
	result := database.DB.Where("user_id = ?", user.ID).Find(&products)

	if result.Error != nil {
		// Handle error fetching products
		renderTemplate(c, "dashboard.html", gin.H{
			"Error":    "Не удалось загрузить ваши товары",
			"Products": []models.Product{}, // Pass empty slice
		})
		return
	}

	renderTemplate(c, "dashboard.html", gin.H{
		"Products": products,
	})
}
