package controllers

import (
	"archive/zip"
	"digital-marketplace/internal/database"
	"digital-marketplace/internal/models"
	"fmt"
	"io"
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
	// Цена может понадобиться в будущем при добавлении функционала цены
	// price := c.PostForm("price")

	// Обеспечим существование директории загрузок
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		renderTemplate(c, "upload.html", gin.H{"Error": "Не удалось создать директорию для загрузок: " + err.Error()})
		return
	}

	// Текущее время для уникальных имен файлов
	timestamp := time.Now().UnixNano()

	// Обработка изображения товара
	var imagePath string
	productImage, err := c.FormFile("product_image")
	if err == nil { // Если изображение предоставлено
		// Создаем уникальное имя для изображения
		imageFilename := fmt.Sprintf("%d_%s", timestamp, filepath.Base(productImage.Filename))
		imageFilePath := filepath.Join(uploadDir, imageFilename)
		webImagePath := "/uploads/" + imageFilename

		// Сохраняем изображение
		if err := c.SaveUploadedFile(productImage, imageFilePath); err != nil {
			renderTemplate(c, "upload.html", gin.H{"Error": "Не удалось сохранить изображение товара: " + err.Error()})
			return
		}

		imagePath = webImagePath
	}

	// Обработка файлов товара
	form, err := c.MultipartForm()
	if err != nil {
		renderTemplate(c, "upload.html", gin.H{"Error": "Ошибка при обработке формы: " + err.Error()})
		return
	}

	files := form.File["product_files"]
	if len(files) == 0 {
		renderTemplate(c, "upload.html", gin.H{"Error": "Необходимо загрузить хотя бы один файл товара"})
		return
	}

	// Создаем временную директорию для загружаемых файлов
	tempDir := filepath.Join(uploadDir, fmt.Sprintf("temp_%d", timestamp))
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		renderTemplate(c, "upload.html", gin.H{"Error": "Не удалось создать временную директорию: " + err.Error()})
		return
	}
	defer os.RemoveAll(tempDir) // Удаляем временную директорию после использования

	// Сохраняем каждый файл во временную директорию
	for _, file := range files {
		// Создаем уникальное имя для каждого файла
		tempFilename := filepath.Join(tempDir, filepath.Base(file.Filename))

		// Сохраняем файл
		if err := c.SaveUploadedFile(file, tempFilename); err != nil {
			renderTemplate(c, "upload.html", gin.H{"Error": "Не удалось сохранить файл: " + err.Error()})
			return
		}
	}

	// Создаем архив
	zipFilename := fmt.Sprintf("%d_product_files.zip", timestamp)
	zipFilePath := filepath.Join(uploadDir, zipFilename)
	webZipPath := "/uploads/" + zipFilename

	// Создаем архив с файлами
	if err := createZipArchive(tempDir, zipFilePath); err != nil {
		renderTemplate(c, "upload.html", gin.H{"Error": "Не удалось создать архив: " + err.Error()})
		return
	}

	// Создаем запись о товаре в БД
	product := models.Product{
		Title:       title,
		Description: description,
		FilePath:    webZipPath,
		ImagePath:   imagePath,
		UserID:      user.ID,
		CreatedAt:   time.Now(),
	}

	result := database.DB.Create(&product)
	if result.Error != nil {
		// При ошибке удаляем созданные файлы
		os.Remove(zipFilePath)
		if imagePath != "" {
			os.Remove(filepath.Join(".", imagePath))
		}
		renderTemplate(c, "upload.html", gin.H{"Error": "Ошибка сохранения товара в базу данных: " + result.Error.Error()})
		return
	}

	// Успешное завершение
	c.Redirect(http.StatusFound, "/dashboard")
}

// Функция для создания zip-архива из файлов в директории
func createZipArchive(sourceDir, destinationPath string) error {
	// Создаем файл архива
	zipFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// Создаем новый zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Обходим все файлы в директории
	err = filepath.Walk(sourceDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем директории
		if info.IsDir() {
			return nil
		}

		// Открываем файл для чтения
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		// Получаем относительный путь для сохранения правильной структуры в архиве
		relPath, err := filepath.Rel(sourceDir, filePath)
		if err != nil {
			return err
		}

		// Создаем файл в архиве
		zipFileEntry, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// Копируем содержимое файла в архив
		_, err = io.Copy(zipFileEntry, file)
		return err
	})

	return err
}
