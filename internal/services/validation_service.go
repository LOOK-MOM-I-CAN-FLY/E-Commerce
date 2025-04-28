package services

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// ValidationService предоставляет функции для валидации пользовательского ввода
type ValidationService struct{}

// NewValidationService создает новый экземпляр ValidationService
func NewValidationService() *ValidationService {
	return &ValidationService{}
}

// Регулярные выражения для валидации
var (
	// Email должен содержать @ и домен с точкой
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// Имя пользователя должно содержать только буквы, цифры и знак подчеркивания
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

	// Проверка допустимых символов в имени тега (буквы, цифры, пробелы, дефисы)
	tagNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-]+$`)

	// Параметры запроса должны содержать только буквы, цифры, пробелы, знаки подчеркивания и дефисы
	queryParamRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\s]+$`)
)

// Список разрешенных MIME-типов для изображений
var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// Список разрешенных расширений файлов продуктов
var AllowedProductFileExtensions = map[string]bool{
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

// Constants
const (
	// Максимальный размер изображения (10 МБ)
	MaxImageSize = 10 * 1024 * 1024

	// Максимальный размер отдельного файла продукта (100 МБ)
	MaxProductFileSize = 100 * 1024 * 1024

	// Максимальное количество файлов для загрузки
	MaxProductFiles = 10

	// Минимальная и максимальная длина имени пользователя
	MinUsernameLen = 3
	MaxUsernameLen = 30

	// Минимальная длина пароля
	MinPasswordLen = 6

	// Максимальная длина названия товара
	MaxTitleLen = 255

	// Максимальная длина описания товара
	MaxDescriptionLen = 5000

	// Максимальная длина имени тега
	MaxTagNameLen = 30
)

// ValidateEmail проверяет корректность email адреса
func (vs *ValidationService) ValidateEmail(email string) (bool, string) {
	email = strings.TrimSpace(email)
	if email == "" {
		return false, "Email не может быть пустым"
	}

	if !emailRegex.MatchString(email) {
		return false, "Некорректный формат email"
	}

	return true, ""
}

// ValidateUsername проверяет корректность имени пользователя
func (vs *ValidationService) ValidateUsername(username string) (bool, string) {
	username = strings.TrimSpace(username)
	if username == "" {
		return false, "Имя пользователя не может быть пустым"
	}

	if len(username) < MinUsernameLen || len(username) > MaxUsernameLen {
		return false, fmt.Sprintf("Имя пользователя должно содержать от %d до %d символов", MinUsernameLen, MaxUsernameLen)
	}

	if !usernameRegex.MatchString(username) {
		return false, "Имя пользователя должно содержать только латинские буквы, цифры и символ подчеркивания"
	}

	return true, ""
}

// ValidatePassword проверяет корректность пароля
func (vs *ValidationService) ValidatePassword(password string) (bool, string) {
	if len(password) < MinPasswordLen {
		return false, fmt.Sprintf("Пароль должен содержать минимум %d символов", MinPasswordLen)
	}

	return true, ""
}

// ValidateTitle проверяет корректность названия товара
func (vs *ValidationService) ValidateTitle(title string) (bool, string) {
	title = strings.TrimSpace(title)
	if title == "" {
		return false, "Название товара не может быть пустым"
	}

	if len(title) > MaxTitleLen {
		return false, fmt.Sprintf("Название товара слишком длинное (максимум %d символов)", MaxTitleLen)
	}

	return true, ""
}

// ValidateDescription проверяет корректность описания товара
func (vs *ValidationService) ValidateDescription(description string) (bool, string) {
	description = strings.TrimSpace(description)

	if len(description) > MaxDescriptionLen {
		return false, fmt.Sprintf("Описание слишком длинное (максимум %d символов)", MaxDescriptionLen)
	}

	return true, ""
}

// ValidatePrice проверяет корректность цены
func (vs *ValidationService) ValidatePrice(price float64) (bool, string) {
	if price < 0 {
		return false, "Цена не может быть отрицательной"
	}

	return true, ""
}

// ValidateTagName проверяет корректность имени тега
func (vs *ValidationService) ValidateTagName(name string) (bool, string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return false, "Имя тега не может быть пустым"
	}

	if len(name) > MaxTagNameLen {
		return false, fmt.Sprintf("Имя тега слишком длинное (максимум %d символов)", MaxTagNameLen)
	}

	if !tagNameRegex.MatchString(name) {
		return false, "Имя тега должно содержать только буквы, цифры, пробелы и дефисы"
	}

	return true, ""
}

// ValidateProductID проверяет корректность ID продукта
func (vs *ValidationService) ValidateProductID(id uint) (bool, string) {
	if id == 0 {
		return false, "ID продукта не может быть равен нулю"
	}

	if id > 1000000 { // Предполагаемый верхний порог
		return false, "ID продукта вне допустимого диапазона"
	}

	return true, ""
}

// ValidateFile проверяет безопасность файла
func (vs *ValidationService) ValidateFile(file *multipart.FileHeader, allowedExtensions map[string]bool, maxSize int64) (bool, string) {
	// Проверка размера файла
	if file.Size > maxSize {
		return false, fmt.Sprintf("Файл слишком большой (максимум %d МБ)", maxSize/(1024*1024))
	}

	// Проверка расширения файла
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		return false, "Файл должен иметь расширение"
	}

	if !allowedExtensions[ext] {
		return false, "Недопустимый тип файла"
	}

	return true, ""
}

// SanitizeFileName очищает имя файла от потенциально опасных символов
func (vs *ValidationService) SanitizeFileName(filename string) string {
	// Получаем только базовое имя файла без пути
	base := filepath.Base(filename)

	// Удаляем все потенциально опасные символы
	// Оставляем только буквы, цифры, точки, дефисы и подчеркивания
	safe := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, base)

	// Если имя пустое после очистки, возвращаем дефолтное имя
	if safe == "" || safe == "." {
		return "file"
	}

	return safe
}

// SanitizeQueryParam очищает параметр запроса
func (vs *ValidationService) SanitizeQueryParam(param string) string {
	// Обрезаем пробелы
	trimmed := strings.TrimSpace(param)

	// Ограничиваем длину
	if len(trimmed) > 100 {
		trimmed = trimmed[:100]
	}

	return trimmed
}

// ValidateQueryParam проверяет безопасность параметра запроса
func (vs *ValidationService) ValidateQueryParam(param string) (bool, string) {
	param = strings.TrimSpace(param)

	// Пустой параметр допустим
	if param == "" {
		return true, ""
	}

	if !queryParamRegex.MatchString(param) {
		return false, "Параметр содержит недопустимые символы"
	}

	return true, ""
}
