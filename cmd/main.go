package main

import (
	"digital-marketplace/internal/controllers"
	"digital-marketplace/internal/database"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Не удалось загрузить .env файл:", err)
	}

	// Checking whether a variable is considered
	log.Println("SMTP_HOST:", os.Getenv("SMTP_HOST"))
	router := gin.Default()

	// Initialize database first
	database.InitDB()

	// Load HTML templates
	router.LoadHTMLGlob("web/templates/*")

	// Serve static files
	router.Static("/static", "./web/static")

	// НЕ обслуживаем /uploads напрямую, используем новые защищенные маршруты
	// router.Static("/uploads", "./uploads") // Закомментировано для безопасности

	// Initialize controllers
	auth := controllers.NewAuthController()
	upload := controllers.NewUploadController()
	buy := controllers.NewBuyController()
	dash := controllers.NewDashboardController()
	prod := controllers.NewProductController()      // New product controller
	cart := controllers.NewCartController()         // Initializing the bucket controller
	order := controllers.NewOrderController()       // Initializing the order controller
	download := controllers.NewDownloadController() // New controller for downloading

	// Public routes (only set login status)
	public := router.Group("/")
	public.Use(controllers.SetLoginStatus()) // Apply middleware to set login status
	{
		public.GET("/", auth.ShowHome) // Use new home handler
		public.GET("/register", auth.ShowRegister)
		public.POST("/register", auth.Register)
		public.GET("/login", auth.ShowLogin)
		public.POST("/login", auth.Login)
		public.GET("/products", prod.ShowProducts) // New products page handler

		// OAuth routes
		public.GET("/auth/github", auth.InitiateGithubLogin)
		public.GET("/auth/github/callback", auth.HandleGithubCallback)

		// Token download route (public, but requires a token)
		public.GET("/download/:token", download.HandleDownload)

		// The route for displaying product images (public)
		public.GET("/images/products/:productID", download.ServeProductImage)
	}

	// Routes requiring authentication
	authenticated := router.Group("/")
	authenticated.Use(controllers.AuthRequired()) // Apply AuthRequired middleware
	{
		authenticated.GET("/logout", auth.Logout)
		authenticated.GET("/dashboard", dash.ShowDashboard)
		authenticated.GET("/upload", upload.ShowUploadPage)
		authenticated.POST("/upload", upload.HandleUpload)
		authenticated.GET("/buy/:productID", buy.ShowBuyPage)
		authenticated.POST("/buy/:productID", buy.HandleBuy)
		authenticated.GET("/profile", auth.ShowProfile)                     // New profile page
		authenticated.POST("/profile/change-password", auth.ChangePassword) // New password change handler

		// Routes for backet
		authenticated.GET("/cart", cart.ShowCart)                       // Show backet
		authenticated.POST("/cart/add/:productID", cart.AddToCart)      // Add product to cart (POST to avoid accidental addition)
		authenticated.POST("/cart/remove/:itemID", cart.RemoveFromCart) // Delete an item from the shopping cart (POST)

		// Routes for placing an order
		authenticated.POST("/checkout", order.Checkout)              // Place an order
		authenticated.GET("/order/success/", order.ShowOrderSuccess) // Successful order page

		// Secure file access routes
		authenticated.GET("/secure-download", download.HandleSecureDownload)       // Через токен
		authenticated.GET("/files/products/:productID", download.ServeProductFile) // Прямой доступ к файлам продуктов
	}

	// Start server
	router.Run(":8080")
}
