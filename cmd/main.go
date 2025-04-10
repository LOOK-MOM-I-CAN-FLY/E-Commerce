package main

import (
	"digital-marketplace/internal/controllers"
	"digital-marketplace/internal/database"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Initialize database first
	database.InitDB()

	// Load HTML templates
	router.LoadHTMLGlob("web/templates/*")

	// Serve static files
	router.Static("/static", "./web/static")
	router.Static("/uploads", "./uploads") // Make sure uploads dir exists

	// Initialize controllers
	auth := controllers.NewAuthController()
	upload := controllers.NewUploadController()
	buy := controllers.NewBuyController()
	dash := controllers.NewDashboardController()
	prod := controllers.NewProductController() // New product controller
	cart := controllers.NewCartController()    // Инициализируем контроллер корзины
	order := controllers.NewOrderController()  // Инициализируем контроллер заказов

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

		// Маршруты для корзины
		authenticated.GET("/cart", cart.ShowCart)                       // Показать корзину
		authenticated.POST("/cart/add/:productID", cart.AddToCart)      // Добавить товар в корзину (POST, чтобы избежать случайного добавления)
		authenticated.POST("/cart/remove/:itemID", cart.RemoveFromCart) // Удалить товар из корзины (POST)

		// Маршруты для оформления заказа
		authenticated.POST("/checkout", order.Checkout)              // Оформить заказ
		authenticated.GET("/order/success/", order.ShowOrderSuccess) // Страница успешного заказа
	}

	// Start server
	router.Run(":8080")
}
