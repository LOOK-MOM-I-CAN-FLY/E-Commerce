package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Загрузка переменных окружения
	err := godotenv.Load()
	if err != nil {
		log.Println("Внимание: .env файл не загружен:", err)
	}

	// Получаем параметры подключения из переменных окружения
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "db" // Используем имя сервиса db по умолчанию в Docker
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "marketplace"
	}

	// Подключение к базе данных
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных:", err)
	}

	// Добавление столбца username к таблице users
	if err := db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(255)").Error; err != nil {
		log.Fatal("Ошибка при добавлении колонки username:", err)
	}

	fmt.Println("Миграция успешно выполнена. Добавлен столбец username в таблицу users.")
}
