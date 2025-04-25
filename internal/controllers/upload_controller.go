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
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UploadController struct{}

func NewUploadController() *UploadController {
	return &UploadController{}
}

// Регулярное выражение для валидации имен тегов:
// Разрешает буквы (Unicode), цифры, пробелы, дефисы.
// Запрещает начинаться/заканчиваться пробелом/дефисом.
// Запрещает несколько пробелов/дефисов подряд.
//var tagNameRegex = regexp.MustCompile(`^[\p{L}\d]+(?:[\s-][\p{L}\d]+)*$`) // Чуть усложним позже, начнем с простого

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

	title := c.PostForm("title")
	description := c.PostForm("description")
	priceStr := c.PostForm("price") // Получаем цену как строку
	// Конвертируем цену в float64
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		renderTemplate(c, "upload.html", gin.H{"Error": "Invalid price format"})
		return
	}

	// Получаем теги из формы
	existingTagIDsStr := c.PostFormArray("existing_tags") // Массив строк ID
	// newTagsStr := c.PostForm("new_tags") // Старый способ
	newTagsListStr := c.PostForm("new_tags_list") // Новый способ: читаем из скрытого поля

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
	// newTagNames := strings.Split(newTagsStr, ",") // Старый способ
	newTagNames := strings.Split(newTagsListStr, ",") // Новый способ
	for _, name := range newTagNames {
		trimmedName := strings.TrimSpace(name)
		if trimmedName == "" {
			continue // Пропускаем пустые строки
		}

		// **ВАЛИДАЦИЯ ИМЕНИ ТЕГА (Серверная)**
		// Простая валидация: не пустое, можно добавить regex позже
		if !isValidTagName(trimmedName) { // Замените на вашу функцию валидации
			renderTemplate(c, "upload.html", gin.H{"Error": fmt.Sprintf("Invalid tag name: '%s'. Only letters, numbers, spaces, hyphens allowed.", trimmedName)})
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
			renderTemplate(c, "upload.html", gin.H{"Error": fmt.Sprintf("Error processing tag '%s': %v", trimmedName, err)})
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
		Price:       price,
		FilePath:    webZipPath,
		ImagePath:   imagePath,
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
		if imagePath != "" {
			// Путь к файлу на диске, а не веб-путь
			localImagePath := filepath.Join(".", imagePath) // Предполагаем, что webImagePath начинается с /uploads/
			os.Remove(localImagePath)
		}
		renderTemplate(c, "upload.html", gin.H{"Error": "Ошибка сохранения товара или тегов: " + err.Error()})
		return
	}
	// --- Конец сохранения в БД ---

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

// Вспомогательная функция для валидации имени тега (простая версия)
func isValidTagName(name string) bool {
	if name == "" {
		return false
	}
	// Простая проверка: разрешаем буквы, цифры, пробелы, дефисы
	// Запрещаем другие символы
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != ' ' && r != '-' {
			return false
		}
	}
	// Можно добавить более сложные проверки (не начинать/заканчивать пробелом/дефисом, не два пробела/дефиса подряд)
	return true
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
