package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	// остальные импорты, например, ваших обработчиков, сервисов и т.п.
)

func main() {
	router := gin.Default()

	// // Группа API, например, /api/…
	// api := router.Group("/api")
	// {
	// 	// Регистрация ваших API-эндпоинтов
	// 	// Пример:
	// 	// api.GET("/products", productHandler.GetProducts)
	// 	// api.POST("/cart/add", orderHandler.AddToCart)
	// }

	// Отдаем главную страницу по "/" вручную
	router.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})

	// Явно прописываем обработчики для остальных страниц,
	// чтобы они не конфликтовали с catch-all маршрутом.
	router.GET("/register.html", func(c *gin.Context) {
		c.File("./web/register.html")
	})
	router.GET("/login.html", func(c *gin.Context) {
		c.File("./web/login.html")
	})
	router.GET("/products.html", func(c *gin.Context) {
		c.File("./web/products.html")
	})
	router.GET("/cart.html", func(c *gin.Context) {
		c.File("./web/cart.html")
	})
	router.GET("/checkout.html", func(c *gin.Context) {
		c.File("./web/checkout.html")
	})

	// Отдаем статические ассеты (CSS, JS, изображения) по другому префиксу,
	// чтобы избежать конфликта с остальными маршрутами.
	if _, err := os.Stat("./web/assets"); err == nil {
		router.Static("/assets", "./web/assets")
	}

	// Если вдруг потребуется отдать любые другие файлы из ./web, можно добавить NoRoute:
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		filePath := "./web" + path
		if _, err := os.Stat(filePath); err == nil {
			c.File(filePath)
		} else {
			c.String(404, "Страница не найдена")
		}
	})

	log.Println("Сервер запущен на http://localhost:8080")
	router.Run(":8080")
}
