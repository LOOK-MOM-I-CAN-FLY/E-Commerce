package controllers

import (
	"digital-marketplace/internal/database"
	"digital-marketplace/internal/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ProductController struct{}

func NewProductController() *ProductController {
	return &ProductController{}
}

// GetTags godoc
// @Summary Get all available tags
// @Description Retrieve a list of all product tags
// @Tags products
// @Accept  json
// @Produce  json
// @Success 200 {array} models.Tag
// @Failure 500 {object} gin.H{"error": string}
// @Router /api/tags [get]
func (pc *ProductController) GetTags(c *gin.Context) {
	var tags []models.Tag
	result := database.DB.Order("name asc").Find(&tags)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
		return
	}
	c.JSON(http.StatusOK, tags)
}

// GetProductsAPI godoc
// @Summary Get products, optionally filtered by tags
// @Description Retrieve a list of products. If 'tags' query parameter is provided (comma-separated), only products matching ALL specified tags are returned.
// @Tags products
// @Accept  json
// @Produce  json
// @Param   tags query string false "Comma-separated list of tag names to filter by"
// @Success 200 {array} models.Product
// @Failure 500 {object} gin.H{"error": string}
// @Router /api/products [get]
func (pc *ProductController) GetProductsAPI(c *gin.Context) {
	var products []models.Product
	tagsQuery := c.Query("tags") // Получаем строку ?tags=tag1,tag2

	if tagsQuery == "" {
		// Если теги не указаны, просто получаем все продукты
		result := database.DB.Find(&products)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}
	} else {
		// Если теги указаны, фильтруем
		tagNames := strings.Split(tagsQuery, ",")
		if len(tagNames) == 0 {
			// Если пустой параметр tags=, возвращаем пустой список
			c.JSON(http.StatusOK, []models.Product{})
			return
		}

		// Используем Raw SQL для сложного запроса с JOIN и HAVING COUNT
		// GORM не очень удобен для таких конструкций
		query := `
			SELECT p.*
			FROM products p
			JOIN product_tags pt ON p.id = pt.product_id
			JOIN tags t ON pt.tag_id = t.id
			WHERE t.name IN (?)
			GROUP BY p.id
			HAVING COUNT(DISTINCT t.id) = ?`

		result := database.DB.Raw(query, tagNames, len(tagNames)).Scan(&products)

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch filtered products"})
			return
		}
		// Если result.RowsAffected == 0, вернется пустой слайс products, что корректно
	}

	c.JSON(http.StatusOK, products)
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

// ShowProductsPage рендерит HTML страницу со всеми продуктами (без фильтрации по тегам здесь)
// Фильтрация будет происходить на фронтенде с помощью запроса к /api/products
func (pc *ProductController) ShowProductsPage(c *gin.Context) {
	// Возможно, здесь не нужно загружать все продукты, если фронтенд все равно их запросит через API?
	// Оставим пока для примера, но для SPA это может быть избыточно.
	var products []models.Product
	database.DB.Find(&products) // Просто получаем все для начальной отрисовки

	// Получаем все теги для отображения фильтров
	var tags []models.Tag
	database.DB.Order("name asc").Find(&tags)

	renderTemplate(c, "products.html", gin.H{
		"AllProducts": products, // Для начального отображения (может быть пустым)
		"AllTags":     tags,     // Для рендеринга фильтров
		// Error handling можно добавить по аналогии с GetProductsAPI
	})
}

// renderTemplate хелпер (должен уже существовать где-то, иначе нужно определить)
// func renderTemplate(c *gin.Context, templateName string, data gin.H) { ... }
// Оставляю заглушку, предполагая, что функция renderTemplate определена глобально или в другом пакете
// Если ее нет, нужно будет ее добавить или использовать c.HTML() напрямую.
