package main

import (
	"fmt"
	"log"

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

	// Подключение к базе данных
	dsn := fmt.Sprintf("host=localhost user=postgres password=postgres dbname=marketplace port=5432 sslmode=disable")
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
