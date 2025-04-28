package controllers

import (
	"digital-marketplace/internal/database"
	"digital-marketplace/internal/models"
	"digital-marketplace/internal/services"
	"html"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Регулярное выражение для проверки параметров можно удалить, т.к. оно перенесено в ValidationService
var alphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9_\- ]+$`)

type ProductController struct {
	validationService *services.ValidationService
}

func NewProductController() *ProductController {
	return &ProductController{
		validationService: services.NewValidationService(),
	}
}

// Проверка безопасности строки запроса (можно удалить, т.к. перенесено в ValidationService)
func sanitizeQueryParam(param string) string {
	// Обрезаем пробелы
	trimmed := strings.TrimSpace(param)

	// Ограничиваем длину
	if len(trimmed) > 100 {
		trimmed = trimmed[:100]
	}

	// Экранируем спецсимволы HTML
	sanitized := html.EscapeString(trimmed)

	return sanitized
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
	tagsParam := c.Query("tags")

	// Используем ValidationService для очистки и валидации параметра
	tagsQuery := pc.validationService.SanitizeQueryParam(tagsParam)
	isValid, _ := pc.validationService.ValidateQueryParam(tagsQuery)

	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недопустимый формат параметра tags"})
		return
	}

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

		// Валидация тегов
		var validTagNames []string
		for _, tag := range tagNames {
			tag = strings.TrimSpace(tag)

			// Используем ValidationService для валидации тега
			isValid, _ := pc.validationService.ValidateTagName(tag)
			if isValid {
				validTagNames = append(validTagNames, tag)
			}
		}

		// Если все теги невалидны, возвращаем пустой список
		if len(validTagNames) == 0 {
			c.JSON(http.StatusOK, []models.Product{})
			return
		}

		// Используем параметризованный запрос для безопасности
		query := `
			SELECT p.*
			FROM products p
			JOIN product_tags pt ON p.id = pt.product_id
			JOIN tags t ON pt.tag_id = t.id
			WHERE t.name IN (?)
			GROUP BY p.id
			HAVING COUNT(DISTINCT t.id) = ?`

		result := database.DB.Raw(query, validTagNames, len(validTagNames)).Scan(&products)

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch filtered products"})
			return
		}
		// Если result.RowsAffected == 0, вернется пустой слайс products, что корректно
	}

	c.JSON(http.StatusOK, products)
}

// ShowProductDetail отображает детальную информацию о продукте по его ID
func (pc *ProductController) ShowProductDetail(c *gin.Context) {
	// Получаем ID продукта из параметров URL
	productIDStr := c.Param("id")

	// Валидация ID
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		renderTemplate(c, "error.html", gin.H{
			"Error": "Неверный ID продукта",
		})
		return
	}

	// Проверка ID через ValidationService
	validID, errMsg := pc.validationService.ValidateProductID(uint(productID))
	if !validID {
		renderTemplate(c, "error.html", gin.H{
			"Error": errMsg,
		})
		return
	}

	// Получаем информацию о продукте
	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		renderTemplate(c, "error.html", gin.H{
			"Error": "Продукт не найден",
		})
		return
	}

	// Получаем теги продукта
	var tags []models.Tag
	if err := database.DB.Joins("JOIN product_tags pt ON pt.tag_id = tags.id").
		Where("pt.product_id = ?", productID).
		Find(&tags).Error; err != nil {
		// Ошибка получения тегов не критична, просто показываем продукт без тегов
		tags = []models.Tag{}
	}

	renderTemplate(c, "product_detail.html", gin.H{
		"Product": product,
		"Tags":    tags,
	})
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

// ShowProductsPage рендерит HTML страницу со всеми продуктами
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
