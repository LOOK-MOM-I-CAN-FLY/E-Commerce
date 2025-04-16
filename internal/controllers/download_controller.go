package controllers

import (
	"digital-marketplace/internal/database"
	"digital-marketplace/internal/models"
	"digital-marketplace/internal/services"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type DownloadController struct {
	fileService *services.FileService
}

func NewDownloadController() *DownloadController {
	return &DownloadController{
		fileService: services.NewFileService(),
	}
}

// HandleDownload обрабатывает запрос на скачивание файла по токену
func (dc *DownloadController) HandleDownload(c *gin.Context) {
	token := c.Param("token")

	// Проверяем существование и действительность токена
	if !dc.fileService.HasValidToken(token) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Недействительный или истекший токен скачивания",
		})
		return
	}

	// Получаем информацию о скачивании
	downloadInfo, err := dc.fileService.GetDownloadInfo(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось получить информацию для скачивания",
		})
		return
	}

	// Проверяем существование файла
	file, err := os.Open(downloadInfo.FilePath)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "Файл не найден",
		})
		return
	}
	defer file.Close()

	// Устанавливаем заголовки для скачивания
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+downloadInfo.FileName)
	c.Header("Content-Type", downloadInfo.ContentType)

	// Отправляем файл
	c.File(downloadInfo.FilePath)

	// Удаляем токен после использования (опционально, можно оставить для повторного скачивания)
	// dc.fileService.DeleteToken(token)

	// Логируем успешное скачивание
	log.Printf("Файл успешно скачан: %s", downloadInfo.FileName)
}

// HandleSecureDownload обрабатывает запрос на защищенное скачивание файла
func (dc *DownloadController) HandleSecureDownload(c *gin.Context) {
	// По умолчанию требуется аутентификация
	user, exists := getUserFromContext(c)
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Требуется авторизация",
		})
		return
	}

	productIDStr := c.Query("product")
	if productIDStr == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Не указан идентификатор продукта",
		})
		return
	}

	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Некорректный идентификатор продукта",
		})
		return
	}

	// Проверим, что пользователь купил этот продукт
	var order models.Order
	result := database.DB.Where("user_id = ?", user.ID).
		Joins("JOIN order_items ON orders.id = order_items.order_id").
		Where("order_items.product_id = ?", productID).
		First(&order)

	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "У вас нет доступа к этому продукту. Пожалуйста, приобретите его сначала.",
		})
		return
	}

	// Получаем информацию о продукте
	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "Продукт не найден",
		})
		return
	}

	// Создаем токен для скачивания
	token, err := dc.fileService.GenerateDownloadToken(uint(productID))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось создать токен для скачивания",
		})
		return
	}

	// Перенаправляем на URL для скачивания
	c.Redirect(http.StatusFound, "/download/"+token)
}

// ServeProductFile обрабатывает запрос на скачивание файла продукта через Go
func (dc *DownloadController) ServeProductFile(c *gin.Context) {
	// Проверяем аутентификацию пользователя
	user, exists := getUserFromContext(c)
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Требуется авторизация",
		})
		return
	}

	// Получаем ID продукта из параметров URL
	productIDStr := c.Param("productID")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Некорректный идентификатор продукта",
		})
		return
	}

	// Проверяем, что пользователь купил этот продукт или является его владельцем
	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "Продукт не найден",
		})
		return
	}

	// Если пользователь не является владельцем, проверяем, купил ли он продукт
	if product.UserID != user.ID {
		var order models.Order
		result := database.DB.Where("user_id = ?", user.ID).
			Joins("JOIN order_items ON orders.id = order_items.order_id").
			Where("order_items.product_id = ?", productID).
			First(&order)

		if result.Error != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "У вас нет доступа к этому продукту. Пожалуйста, приобретите его сначала.",
			})
			return
		}
	}

	// Проверяем существование файла
	filePath := strings.TrimPrefix(product.FilePath, "/")
	fullPath := filepath.Join(".", filePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "Файл продукта не найден",
		})
		return
	}

	// Устанавливаем заголовки для скачивания
	fileName := filepath.Base(product.FilePath)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", dc.fileService.GuessContentType(filePath))

	// Отправляем файл
	c.File(fullPath)

	// Логируем успешное скачивание
	log.Printf("Пользователь %d скачал файл продукта %d: %s", user.ID, productID, fileName)
}

// ServeProductImage обрабатывает запрос на отображение изображения продукта
func (dc *DownloadController) ServeProductImage(c *gin.Context) {
	// Получаем ID продукта из параметров URL
	productIDStr := c.Param("productID")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Некорректный идентификатор продукта",
		})
		return
	}

	// Получаем информацию о продукте
	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "Продукт не найден",
		})
		return
	}

	// Определяем, какой путь к изображению использовать
	var imagePath string
	if product.ImagePath != "" {
		imagePath = product.ImagePath
	} else if product.FilePath != "" && (strings.HasSuffix(strings.ToLower(product.FilePath), ".jpg") ||
		strings.HasSuffix(strings.ToLower(product.FilePath), ".jpeg") ||
		strings.HasSuffix(strings.ToLower(product.FilePath), ".png") ||
		strings.HasSuffix(strings.ToLower(product.FilePath), ".gif") ||
		strings.HasSuffix(strings.ToLower(product.FilePath), ".webp")) {
		imagePath = product.FilePath
	} else {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "Изображение продукта не найдено",
		})
		return
	}

	// Проверяем существование файла
	imagePath = strings.TrimPrefix(imagePath, "/")
	fullPath := filepath.Join(".", imagePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "Файл изображения не найден",
		})
		return
	}

	// Отправляем изображение
	c.File(fullPath)
}
