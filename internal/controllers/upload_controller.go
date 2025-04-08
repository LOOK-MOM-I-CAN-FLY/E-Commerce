package controllers

import (
	"digital-marketplace/internal/database"
	"digital-marketplace/internal/models"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type UploadController struct{}

func NewUploadController() *UploadController {
	return &UploadController{}
}

func (uc *UploadController) ShowUploadPage(c *gin.Context) {
	// Use renderTemplate, assuming upload.html doesn't need specific data beyond login status
	renderTemplate(c, "upload.html", gin.H{})
}

func (uc *UploadController) HandleUpload(c *gin.Context) {
	// Get user from context (set by AuthRequired middleware)
	user, exists := getUserFromContext(c)
	if !exists {
		// Should be handled by middleware, but redirect as fallback
		c.Redirect(http.StatusFound, "/login")
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")
	file, err := c.FormFile("file")
	if err != nil {
		// Render error on the upload page
		renderTemplate(c, "upload.html", gin.H{"Error": "Ошибка загрузки файла: " + err.Error()})
		return
	}

	// Ensure uploads directory exists
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		renderTemplate(c, "upload.html", gin.H{"Error": "Не удалось создать директорию для загрузок: " + err.Error()})
		return
	}

	// Create a unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	filePath := filepath.Join(uploadDir, filename)
	webPath := "/uploads/" + filename

	// Save the file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		renderTemplate(c, "upload.html", gin.H{"Error": "Не удалось сохранить файл: " + err.Error()})
		return
	}

	// Create product entry in DB
	product := models.Product{
		Title:       title,
		Description: description,
		FilePath:    webPath,
		UserID:      user.ID,
		CreatedAt:   time.Now(),
	}

	result := database.DB.Create(&product)
	if result.Error != nil {
		// Ideally, clean up saved file if DB save fails
		// os.Remove(filePath)
		renderTemplate(c, "upload.html", gin.H{"Error": "Ошибка сохранения товара в базу данных: " + result.Error.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/dashboard") // Use StatusFound
}
