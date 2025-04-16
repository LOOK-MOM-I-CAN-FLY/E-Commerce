package controllers

import (
	"digital-marketplace/internal/database"
	"digital-marketplace/internal/models"

	"github.com/gin-gonic/gin"
)

type ProductController struct{}

func NewProductController() *ProductController {
	return &ProductController{}
}

// ShowProducts displays the page with all products
func (pc *ProductController) ShowProducts(c *gin.Context) {
	var products []models.Product // Assuming you have a Product model in internal/models
	result := database.DB.Find(&products)

	if result.Error != nil {
		// Handle error - maybe show an error page or log it
		renderTemplate(c, "products.html", gin.H{
			"Error":       "Не удалось загрузить товары",
			"AllProducts": []models.Product{}, // Pass empty slice on error
		})
		return
	}

	renderTemplate(c, "products.html", gin.H{
		"AllProducts": products,
	})
}
