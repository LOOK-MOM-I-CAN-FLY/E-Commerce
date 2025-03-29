package handlers

import (
	"fmt"
	"net/http"

	"github.com/LOOK-MOM-I-CAN-FLY/E-Commerce/internal/models"
	"github.com/LOOK-MOM-I-CAN-FLY/E-Commerce/internal/service"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService *service.OrderService
}

// Временное хранилище корзин
var carts = make(map[string][]string)
var anonymousCart []string // Корзина для неавторизованных пользователей

func NewOrderHandler(os *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: os}
}

// POST /api/cart/add
func (oh *OrderHandler) AddToCart(c *gin.Context) {
	// Получаем пользователя из сессии, если он доступен
	userEmail, err := c.Cookie("session_token")

	var payload struct {
		ProductID string `json:"product_id"`
	}
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	if err != nil || userEmail == "" {
		// Если пользователь не аутентифицирован, используем анонимную корзину
		anonymousCart = append(anonymousCart, payload.ProductID)
		c.JSON(http.StatusOK, gin.H{"message": "Товар добавлен в корзину", "anonymous": true})
	} else {
		// Пользователь аутентифицирован, добавляем в его корзину
		carts[userEmail] = append(carts[userEmail], payload.ProductID)
		c.JSON(http.StatusOK, gin.H{"message": "Товар добавлен в корзину"})
	}
}

// GET /api/cart
func (oh *OrderHandler) ViewCart(c *gin.Context) {
	userEmail, err := c.Cookie("session_token")

	// Получаем соответствующую корзину
	var cartItems []string
	if err != nil || userEmail == "" {
		cartItems = anonymousCart
	} else {
		cartItems = carts[userEmail]
	}

	// Получаем детали продуктов
	products := []models.Product{}
	for _, productID := range cartItems {
		product, ok := oh.orderService.GetProductByID(productID)
		if ok {
			products = append(products, product)
		}
	}

	c.JSON(http.StatusOK, products)
}

// POST /api/checkout
func (oh *OrderHandler) Checkout(c *gin.Context) {
	userEmail, _ := c.Cookie("session_token")

	var payload struct {
		Email string `json:"email"`
		Name  string `json:"name,omitempty"`
	}
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	// Определяем, какую корзину использовать
	var productsToCheckout []string
	if userEmail == "" {
		productsToCheckout = anonymousCart
	} else {
		productsToCheckout = carts[userEmail]
	}

	if len(productsToCheckout) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Корзина пуста"})
		return
	}

	// Обработка заказа
	err := oh.orderService.Checkout(userEmail, payload.Email, productsToCheckout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Ошибка оформления заказа: %v", err)})
		return
	}

	// Очищаем корзину
	if userEmail == "" {
		anonymousCart = []string{}
	} else {
		carts[userEmail] = []string{}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Заказ успешно оформлен"})
}
