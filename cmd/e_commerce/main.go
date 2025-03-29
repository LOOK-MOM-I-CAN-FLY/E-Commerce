package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/LOOK-MOM-I-CAN-FLY/E-Commerce/internal/handlers"
	"github.com/LOOK-MOM-I-CAN-FLY/E-Commerce/internal/repository"
	"github.com/LOOK-MOM-I-CAN-FLY/E-Commerce/internal/service"
)

func main() {
	// Ярко выделяем, что запущена новая версия
	fmt.Println("\n\n========== ЗАПУЩЕНА НОВАЯ ВЕРСИЯ FRAMESTORE (E-Commerce_2) ==========\n\n")

	// Определяем, где мы находимся
	wd, _ := os.Getwd()
	fmt.Println("Рабочая директория:", wd)

	// Проверка существования файла React
	if _, err := os.Stat("./web/build/index.html"); os.IsNotExist(err) {
		fmt.Println("ОШИБКА: Файл ./web/build/index.html не найден!")
	} else {
		fmt.Println("Файл ./web/build/index.html найден, React-приложение будет обслуживаться")
	}

	router := gin.Default()

	// Настройка CORS с использованием готовой функции Gin
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Добавим middleware для логирования всех запросов
	router.Use(func(c *gin.Context) {
		fmt.Printf("Запрос: %s %s\n", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

	// Инициализация тестовой базы данных в памяти
	productRepo := repository.NewProductRepository()
	orderRepo := repository.NewOrderRepository()

	// Инициализация сервисов и обработчиков
	orderService := service.NewOrderService(productRepo, orderRepo)
	orderHandler := handlers.NewOrderHandler(orderService)

	productService := service.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	// Главный маршрут - перенаправляем на React-приложение
	router.GET("/", func(c *gin.Context) {
		fmt.Println("Запрос к корневому маршруту / - отправляем React-приложение (index.html)")
		// Добавляем заголовок, чтобы браузер не кэшировал
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.File("./web/build/index.html")
	})

	// Перенаправления для старых HTML-страниц
	router.GET("/index.html", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/")
	})

	router.GET("/register.html", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/register")
	})

	router.GET("/login.html", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/login")
	})

	router.GET("/products.html", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/products")
	})

	router.GET("/cart.html", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/cart")
	})

	router.GET("/checkout.html", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/checkout")
	})

	// API endpoints
	api := router.Group("/api")
	{
		// Продукты
		api.GET("/products", productHandler.GetProducts)

		// Корзина
		api.POST("/cart/add", orderHandler.AddToCart)
		api.GET("/cart", orderHandler.ViewCart)

		// Оформление заказа
		api.POST("/checkout", orderHandler.Checkout)

		// Аутентификация (заглушка для тестирования)
		api.POST("/login", func(c *gin.Context) {
			// Упрощенная имитация аутентификации
			cookie := &http.Cookie{
				Name:     "session_token",
				Value:    "test@example.com",
				Path:     "/",
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			}
			http.SetCookie(c.Writer, cookie)
			c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Успешный вход"})
		})
	}

	// Статические файлы React-приложения
	router.Static("/static", "./web/build/static")
	router.StaticFile("/favicon.ico", "./web/build/favicon.ico")
	router.StaticFile("/manifest.json", "./web/build/manifest.json")
	router.StaticFile("/logo192.png", "./web/build/logo192.png")
	router.StaticFile("/logo512.png", "./web/build/logo512.png")

	// Маршрут для ReactRouter
	router.NoRoute(func(c *gin.Context) {
		// Проверяем, является ли запрос API
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}

		// Добавляем заголовок, чтобы браузер не кэшировал
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		fmt.Println("Запрос к неизвестному маршруту", c.Request.URL.Path, "- отправляем React-приложение (index.html)")
		// Все остальные запросы обрабатываем через React Router
		c.File("./web/build/index.html")
	})

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("НОВАЯ ВЕРСИЯ FrameStore запущена на http://localhost:%s\n", port)
	router.Run(":" + port)
}
