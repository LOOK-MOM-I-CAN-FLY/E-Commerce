package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Проверяем, что текущая директория содержит папку uploads
	if _, err := os.Stat("../../uploads"); os.IsNotExist(err) {
		fmt.Println("Директория uploads не найдена. Запустите скрипт из корневой директории проекта.")
		return
	}

	fmt.Println("Начинаем очистку директории uploads...")

	uploadDir := "../../uploads"

	// Получаем список всех файлов в директории
	files, err := os.ReadDir(uploadDir)
	if err != nil {
		fmt.Printf("Ошибка при чтении директории: %v\n", err)
		return
	}

	// Счетчики для статистики
	var totalFiles, removedFiles, keptFiles int
	var totalSize, removedSize int64

	// Проходимся по всем файлам
	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			// Если это директория с temp_ в начале имени, удаляем её
			if strings.HasPrefix(fileInfo.Name(), "temp_") {
				fullPath := filepath.Join(uploadDir, fileInfo.Name())
				err := os.RemoveAll(fullPath)
				if err != nil {
					fmt.Printf("Не удалось удалить временную директорию %s: %v\n", fileInfo.Name(), err)
				} else {
					fmt.Printf("Удалена временная директория: %s\n", fileInfo.Name())
					removedFiles++
				}
			}
			continue
		}

		// Получаем информацию о файле
		file, err := fileInfo.Info()
		if err != nil {
			fmt.Printf("Не удалось получить информацию о файле %s: %v\n", fileInfo.Name(), err)
			continue
		}

		// Подсчитываем общее количество файлов и их размер
		totalFiles++
		totalSize += file.Size()

		// Проверяем, является ли файл временным
		fileName := file.Name()

		// Сохраняем все файлы с расширениями изображений и архивы продуктов
		if fileName == ".keep" ||
			strings.HasSuffix(fileName, "_product_files.zip") ||
			// Все изображения разных форматов нужно сохранить
			strings.HasSuffix(strings.ToLower(fileName), ".jpg") ||
			strings.HasSuffix(strings.ToLower(fileName), ".jpeg") ||
			strings.HasSuffix(strings.ToLower(fileName), ".png") ||
			strings.HasSuffix(strings.ToLower(fileName), ".gif") ||
			strings.HasSuffix(strings.ToLower(fileName), ".webp") {
			keptFiles++
			fmt.Printf("Оставлен файл: %s\n", fileName)
		} else {
			// Удаляем временный файл
			fullPath := filepath.Join(uploadDir, fileName)
			err := os.Remove(fullPath)
			if err != nil {
				fmt.Printf("Не удалось удалить файл %s: %v\n", fileName, err)
			} else {
				fmt.Printf("Удален файл: %s (размер: %d байт)\n", fileName, file.Size())
				removedFiles++
				removedSize += file.Size()
			}
		}
	}

	// Выводим статистику
	fmt.Printf("\nСтатистика очистки:\n")
	fmt.Printf("Всего файлов: %d (размер: %d байт)\n", totalFiles, totalSize)
	fmt.Printf("Удалено файлов: %d (размер: %d байт)\n", removedFiles, removedSize)
	fmt.Printf("Оставлено файлов: %d (размер: %d байт)\n", keptFiles, totalSize-removedSize)

	fmt.Println("Очистка завершена!")
}
