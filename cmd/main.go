package main

import (
	"digital-marketplace/internal/controllers"
	"digital-marketplace/internal/database"
	"log"
	"os"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Failed to load .env file:", err)
	}

	// Check if environment variable is read correctly
	log.Println("SMTP_HOST:", os.Getenv("SMTP_HOST"))
	router := gin.Default()

	// Initialize the database
	database.InitDB()

	// Load HTML templates with дополнительными функциями
	router.SetFuncMap(template.FuncMap{
		"subtract": func(a, b float64) float64 {
			return a - b
		},
	})
	router.LoadHTMLGlob("web/templates/*")

	// Serve static files
	router.Static("/static", "./web/static")

	// Do NOT serve /uploads directly, use new protected routes instead
	// router.Static("/uploads", "./uploads") // Commented out for security

	// Добавляем эндпоинт для проверки работоспособности (health check)
	router.GET("/health", func(c *gin.Context) {
		c.String(200, "OK")
	})

	// Initialize controllers
	auth := controllers.NewAuthController()
	upload := controllers.NewUploadController()
	buy := controllers.NewBuyController()
	prod := controllers.NewProductController()      // Product controller
	cart := controllers.NewCartController()         // Cart controller
	order := controllers.NewOrderController()       // Order controller
	download := controllers.NewDownloadController() // Download controller

	// Public routes (only set login status)
	public := router.Group("/")
	public.Use(controllers.SetLoginStatus()) // Middleware to set login status
	{
		public.GET("/", auth.ShowHome) // Homepage handler
		public.GET("/register", auth.ShowRegister)
		public.POST("/register", auth.Register)
		public.GET("/login", auth.ShowLogin)
		public.POST("/login", auth.Login)
		public.GET("/products", prod.ShowProductsPage) // Новый вариант, рендерит HTML страницу

		// OAuth routes
		public.GET("/auth/github", auth.InitiateGithubLogin)
		public.GET("/auth/github/callback", auth.HandleGithubCallback)

		// Route to download with token (public but token-protected)
		public.GET("/download/:token", download.HandleDownload)

		// Route to serve product images (public)
		public.GET("/images/products/:productID", download.ServeProductImage)
	}

	// Routes requiring authentication
	authenticated := router.Group("/")
	authenticated.Use(controllers.AuthRequired()) // Middleware to require authentication
	{
		authenticated.GET("/logout", auth.Logout)
		authenticated.GET("/upload", upload.ShowUploadPage)
		authenticated.POST("/upload", upload.HandleUpload)
		authenticated.GET("/buy/:productID", buy.ShowBuyPage)
		authenticated.POST("/buy/:productID", buy.HandleBuy)
		authenticated.GET("/profile", auth.ShowProfile)                     // Profile page
		authenticated.POST("/profile/change-password", auth.ChangePassword) // Change password handler
		authenticated.POST("/earn-money", auth.EarnMoney)                   // Маршрут для заработка денег

		// Cart routes
		authenticated.GET("/cart", cart.ShowCart)                       // Show cart
		authenticated.POST("/cart/add/:productID", cart.AddToCart)      // Add product to cart (POST to avoid accidental adds)
		authenticated.POST("/cart/remove/:itemID", cart.RemoveFromCart) // Remove product from cart (POST)

		// Checkout routes
		authenticated.POST("/checkout", order.Checkout)              // Checkout handler
		authenticated.GET("/order/success/", order.ShowOrderSuccess) // Order success page

		// Protected routes for file access
		authenticated.GET("/secure-download", download.HandleSecureDownload)       // Download via token
		authenticated.GET("/files/products/:productID", download.ServeProductFile) // Direct access to product files
	}

	// API routes (JSON endpoints)
	api := router.Group("/api")
	{
		// Добавляем маршруты для API продуктов и тегов
		api.GET("/products", prod.GetProductsAPI) // Получение списка продуктов (JSON)
		api.GET("/tags", prod.GetTags)            // Получение списка тегов (JSON)
		// Сюда можно добавить другие API эндпоинты в будущем
	}

	// Start server
	log.Println("Server started on port http://localhost:8080")
	router.Run(":8080")
}
