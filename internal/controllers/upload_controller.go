package controllers

import (
	"archive/zip"
	"digital-marketplace/internal/database"
	"digital-marketplace/internal/models"
	"digital-marketplace/internal/services"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Список разрешенных MIME-типов для изображений
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// Максимальный размер изображения (10 МБ)
const maxImageSize = 10 * 1024 * 1024

// Максимальный размер отдельного файла продукта (100 МБ)
const maxProductFileSize = 100 * 1024 * 1024

// Максимальное количество файлов для загрузки
const maxProductFiles = 10

// Список разрешенных расширений файлов продуктов
var allowedProductFileExtensions = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".txt":  true,
	".rtf":  true,
	".zip":  true,
	".rar":  true,
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".mp3":  true,
	".mp4":  true,
	".avi":  true,
	".mov":  true,
	".psd":  true,
	".ai":   true,
	".eps":  true,
	".svg":  true,
}

type UploadController struct {
	validationService *services.ValidationService
}

func NewUploadController() *UploadController {
	return &UploadController{
		validationService: services.NewValidationService(),
	}
}

// Регулярное выражение для валидации имен тегов:
// Разрешает буквы (Unicode), цифры, пробелы, дефисы.
// Запрещает начинаться/заканчиваться пробелом/дефисом.
// Запрещает несколько пробелов/дефисов подряд.
//var tagNameRegex = regexp.MustCompile(`^[\p{L}\d]+(?:[\s-][\p{L}\d]+)*$`) // Чуть усложним позже, начнем с простого

// Проверка безопасности файла перенесена в ValidationService

// Проверка безопасности имени файла перенесена в ValidationService

func (uc *UploadController) ShowUploadPage(c *gin.Context) {
	// Загружаем существующие теги для передачи в шаблон
	var existingTags []models.Tag
	dbResult := database.DB.Order("name asc").Find(&existingTags)
	if dbResult.Error != nil {
		// Логгируем ошибку, но продолжаем рендерить страницу,
		// возможно, без списка существующих тегов
		fmt.Println("Error fetching existing tags:", dbResult.Error)
		renderTemplate(c, "upload.html", gin.H{"Error": "Could not load existing tags."})
		return
	}

	// Используем renderTemplate, передавая существующие теги
	renderTemplate(c, "upload.html", gin.H{
		"ExistingTags": existingTags, // Передаем теги в шаблон
	})
}

func (uc *UploadController) HandleUpload(c *gin.Context) {
	// Get user from context (set by AuthRequired middleware)
	user, exists := getUserFromContext(c)
	if !exists {
		// Should be handled by middleware, but redirect as fallback
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Валидация полей формы
	title := strings.TrimSpace(c.PostForm("title"))
	description := strings.TrimSpace(c.PostForm("description"))
	priceStr := strings.TrimSpace(c.PostForm("price"))

	// Проверка названия товара
	if valid, errMsg := uc.validationService.ValidateTitle(title); !valid {
		renderTemplate(c, "upload.html", gin.H{"Error": errMsg})
		return
	}

	// Проверка описания товара
	if valid, errMsg := uc.validationService.ValidateDescription(description); !valid {
		renderTemplate(c, "upload.html", gin.H{
			"Error": errMsg,
			"Title": title,
		})
		return
	}

	// Конвертируем и проверяем цену
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		renderTemplate(c, "upload.html", gin.H{
			"Error":       "Неверный формат цены. Введите числовое значение",
			"Title":       title,
			"Description": description,
		})
		return
	}

	if valid, errMsg := uc.validationService.ValidatePrice(price); !valid {
		renderTemplate(c, "upload.html", gin.H{
			"Error":       errMsg,
			"Title":       title,
			"Description": description,
		})
		return
	}

	// Получаем теги из формы
	existingTagIDsStr := c.PostFormArray("existing_tags") // Массив строк ID
	newTagsListStr := c.PostForm("new_tags_list")         // Новый способ: читаем из скрытого поля

	// --- Обработка тегов ---
	var tagIDs []int
	processedTagNames := make(map[string]bool) // Для избежания дубликатов по имени

	// 1. Обрабатываем существующие выбранные теги
	for _, idStr := range existingTagIDsStr {
		id, err := strconv.Atoi(idStr)
		if err == nil {
			tagIDs = append(tagIDs, id)
			// Можно добавить получение имени тега по ID и добавление в processedTagNames,
			// чтобы новые теги с таким же именем не дублировались
		}
	}

	// 2. Обрабатываем новые теги из newTagsListStr
	newTagNames := strings.Split(newTagsListStr, ",") // Новый способ
	for _, name := range newTagNames {
		trimmedName := strings.TrimSpace(name)
		if trimmedName == "" {
			continue // Пропускаем пустые строки
		}

		// **ВАЛИДАЦИЯ ИМЕНИ ТЕГА (Серверная)** с использованием ValidationService
		if valid, errMsg := uc.validationService.ValidateTagName(trimmedName); !valid {
			renderTemplate(c, "upload.html", gin.H{
				"Error":       errMsg,
				"Title":       title,
				"Description": description,
			})
			return
		}

		// Проверяем, не обрабатывали ли уже тег с таким именем (из существующих или новых)
		lowerCaseName := strings.ToLower(trimmedName)
		if processedTagNames[lowerCaseName] {
			continue
		}

		var tag models.Tag
		// Ищем или создаем тег (регистронезависимо)
		err := database.DB.Where("lower(name) = ?", lowerCaseName).FirstOrCreate(&tag, models.Tag{Name: trimmedName}).Error
		if err != nil {
			renderTemplate(c, "upload.html", gin.H{
				"Error":       fmt.Sprintf("Ошибка обработки тега '%s': %v", trimmedName, err),
				"Title":       title,
				"Description": description,
			})
			return
		}

		tagIDs = append(tagIDs, tag.ID)
		processedTagNames[lowerCaseName] = true
	}

	// Удаляем дубликаты ID, если они могли появиться
	tagIDs = uniqueInts(tagIDs)
	// --- Конец обработки тегов ---

	// Обеспечим существование директории загрузок
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		renderTemplate(c, "upload.html", gin.H{
			"Error":       "Не удалось создать директорию для загрузок: " + err.Error(),
			"Title":       title,
			"Description": description,
		})
		return
	}

	// Текущее время для уникальных имен файлов
	timestamp := time.Now().UnixNano()

	// Обработка загруженного изображения товара
	image, err := c.FormFile("image")
	if err != nil {
		renderTemplate(c, "upload.html", gin.H{
			"Error":       "Ошибка при загрузке изображения товара",
			"Title":       title,
			"Description": description,
		})
		return
	}

	// Валидация изображения
	if valid, errMsg := uc.validationService.ValidateFile(image, false); !valid {
		renderTemplate(c, "upload.html", gin.H{
			"Error":       fmt.Sprintf("Проблема с изображением товара: %s", errMsg),
			"Title":       title,
			"Description": description,
		})
		return
	}

	// Обработка файлов продукта
	form, err := c.MultipartForm()
	if err != nil {
		renderTemplate(c, "upload.html", gin.H{
			"Error":       "Ошибка при обработке формы",
			"Title":       title,
			"Description": description,
		})
		return
	}

	files := form.File["files"]
	if len(files) > maxProductFiles {
		renderTemplate(c, "upload.html", gin.H{
			"Error":       fmt.Sprintf("Превышено максимальное количество файлов (%d)", maxProductFiles),
			"Title":       title,
			"Description": description,
		})
		return
	}

	// Валидация каждого файла продукта
	for _, file := range files {
		if valid, errMsg := uc.validationService.ValidateFile(file, true); !valid {
			renderTemplate(c, "upload.html", gin.H{
				"Error":       fmt.Sprintf("Проблема с файлом %s: %s", file.Filename, errMsg),
				"Title":       title,
				"Description": description,
			})
			return
		}
	}

	// Создаем временную директорию для загружаемых файлов
	tempDir := filepath.Join(uploadDir, fmt.Sprintf("temp_%d", timestamp))
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		renderTemplate(c, "upload.html", gin.H{
			"Error":       "Не удалось создать временную директорию: " + err.Error(),
			"Title":       title,
			"Description": description,
		})
		return
	}
	defer os.RemoveAll(tempDir) // Удаляем временную директорию после использования

	// Сохраняем каждый файл во временную директорию
	for _, file := range files {
		// Очистка имени файла с использованием ValidationService
		safeFilename := uc.validationService.SanitizeFileName(file.Filename)

		// Создаем уникальное имя для каждого файла
		tempFilename := filepath.Join(tempDir, safeFilename)

		// Сохраняем файл
		if err := c.SaveUploadedFile(file, tempFilename); err != nil {
			renderTemplate(c, "upload.html", gin.H{
				"Error":       "Не удалось сохранить файл: " + err.Error(),
				"Title":       title,
				"Description": description,
			})
			return
		}
	}

	// Создаем архив
	zipFilename := fmt.Sprintf("%d_product_files.zip", timestamp)
	zipFilePath := filepath.Join(uploadDir, zipFilename)
	webZipPath := "/uploads/" + zipFilename

	// Создаем архив с файлами
	if err := createZipArchive(tempDir, zipFilePath); err != nil {
		renderTemplate(c, "upload.html", gin.H{
			"Error":       "Не удалось создать архив: " + err.Error(),
			"Title":       title,
			"Description": description,
		})
		return
	}

	// Создаем запись о товаре в БД
	product := models.Product{
		Title:       title,
		Description: description,
		Price:       price,
		FilePath:    webZipPath,
		UserID:      user.ID,
		CreatedAt:   time.Now(),
	}

	// --- Сохранение в БД в транзакции ---
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Создаем продукт
		if err := tx.Create(&product).Error; err != nil {
			return err // Возвращаем ошибку для отката транзакции
		}

		// 2. Создаем связи с тегами
		if len(tagIDs) > 0 {
			var productTags []models.ProductTag
			for _, tagID := range tagIDs {
				productTags = append(productTags, models.ProductTag{ProductID: product.ID, TagID: tagID})
			}
			if err := tx.Create(&productTags).Error; err != nil {
				return err // Возвращаем ошибку для отката транзакции
			}
		}

		return nil // Все успешно, коммитим транзакцию
	})

	if err != nil {
		// Ошибка транзакции: удаляем созданные файлы и показываем ошибку
		os.Remove(zipFilePath)
		renderTemplate(c, "upload.html", gin.H{
			"Error":       "Ошибка сохранения товара или тегов: " + err.Error(),
			"Title":       title,
			"Description": description,
		})
		return
	}
	// --- Конец сохранения в БД ---

	// Успешное завершение
	c.Redirect(http.StatusFound, "/profile")
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

// Вспомогательная функция для удаления дубликатов из слайса int
func uniqueInts(intSlice []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
