package controllers

import (
	"digital-marketplace/internal/services"
	"log"
	"net/http"
	"os"

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

	// Создаем токен для скачивания
	token, err := dc.fileService.GenerateDownloadToken(user.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось создать токен для скачивания",
		})
		return
	}

	// Перенаправляем на URL для скачивания
	c.Redirect(http.StatusFound, "/download/"+token)
}
