package services

import (
	"crypto/rand"
	"digital-marketplace/internal/database"
	"digital-marketplace/internal/models"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileService предоставляет методы для работы с файлами
type FileService struct{}

// NewFileService создает новый экземпляр сервиса файлов
func NewFileService() *FileService {
	return &FileService{}
}

// DownloadInfo содержит информацию для безопасного скачивания файла
type DownloadInfo struct {
	FileName    string
	ContentType string
	FilePath    string
	ExpireTime  time.Time
}

var activeDownloads = make(map[string]DownloadInfo)

// GenerateDownloadToken создает временный токен для скачивания файла
func (fs *FileService) GenerateDownloadToken(productID uint) (string, error) {
	// Получаем информацию о продукте
	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		return "", errors.New("продукт не найден")
	}

	// Проверяем существование файла
	filePath := strings.TrimPrefix(product.FilePath, "/")
	if !strings.HasPrefix(filePath, "uploads/") {
		filePath = "uploads/" + filepath.Base(filePath)
	}

	fullPath := "./" + filePath
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", errors.New("файл продукта не существует")
	}

	// Генерируем случайный токен
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", errors.New("не удалось создать токен")
	}
	token := hex.EncodeToString(tokenBytes)

	// Сохраняем информацию о скачивании
	downloadInfo := DownloadInfo{
		FileName:    filepath.Base(product.FilePath),
		ContentType: fs.GuessContentType(filePath),
		FilePath:    fullPath,
		ExpireTime:  time.Now().Add(24 * time.Hour), // Токен действителен 24 часа
	}

	activeDownloads[token] = downloadInfo
	return token, nil
}

// HasValidToken проверяет действительность токена скачивания
func (fs *FileService) HasValidToken(token string) bool {
	info, exists := activeDownloads[token]
	if !exists {
		return false
	}

	// Проверяем срок действия
	if time.Now().After(info.ExpireTime) {
		delete(activeDownloads, token) // Удаляем просроченный токен
		return false
	}

	return true
}

// GetDownloadInfo возвращает информацию о скачивании по токену
func (fs *FileService) GetDownloadInfo(token string) (DownloadInfo, error) {
	info, exists := activeDownloads[token]
	if !exists {
		return DownloadInfo{}, errors.New("недействительный токен скачивания")
	}

	// Проверяем срок действия
	if time.Now().After(info.ExpireTime) {
		delete(activeDownloads, token) // Удаляем просроченный токен
		return DownloadInfo{}, errors.New("срок действия токена истек")
	}

	return info, nil
}

// GenerateDownloadURL создает полный URL для скачивания файла с использованием токена
func (fs *FileService) GenerateDownloadURL(token string, baseURL string) string {
	return fmt.Sprintf("%s/download/%s", baseURL, token)
}

// DeleteToken удаляет токен после использования
func (fs *FileService) DeleteToken(token string) {
	delete(activeDownloads, token)
}

// GuessContentType определяет тип содержимого файла
func (fs *FileService) GuessContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	case ".mp3":
		return "audio/mpeg"
	case ".mp4":
		return "video/mp4"
	default:
		return "application/octet-stream"
	}
}

// GenerateSecureURL создает защищенный временный URL для скачивания с подписью
func GenerateSecureURL(productID uint, userID uint, baseURL string) (string, error) {
	// Генерируем случайный ключ
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Создаем подпись с информацией о пользователе, продукте и времени
	timestamp := time.Now().Unix()
	signature := fmt.Sprintf("%d:%d:%d", userID, productID, timestamp)

	// Кодируем в base64
	combined := append(randomBytes, []byte(signature)...)
	token := base64.URLEncoding.EncodeToString(combined)

	return fmt.Sprintf("%s/secure-download/%s", baseURL, token), nil
}
