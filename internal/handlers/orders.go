package handlers

import (
	"net/http"

	"github.com/LOOK-MOM-I-CAN-FLY/E-Commerce/internal/service"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService *service.OrderService
}

// Временное хранилище корзин: сопоставление email пользователя → срез идентификаторов товаров
var carts = make(map[string][]string)

func NewOrderHandler(os *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: os}
}

// POST /api/cart/add
func (oh *OrderHandler) AddToCart(c *gin.Context) {
	// Здесь для простоты предполагаем, что аутентификация реализована через cookie "session_token"
	userEmail, err := c.Cookie("session_token")
	if err != nil {
		c.String(http.StatusUnauthorized, "Пользователь не аутентифицирован")
		return
	}
	var payload struct {
		ProductID string `json:"product_id"`
	}
	if err := c.BindJSON(&payload); err != nil {
		c.String(http.StatusBadRequest, "Неверные данные")
		return
	}
	carts[userEmail] = append(carts[userEmail], payload.ProductID)
	c.String(http.StatusOK, "Товар добавлен в корзину")
}

// GET /api/cart
func (oh *OrderHandler) ViewCart(c *gin.Context) {
	userEmail, err := c.Cookie("session_token")
	if err != nil {
		c.String(http.StatusUnauthorized, "Пользователь не аутентифицирован")
		return
	}
	// Здесь можно было бы вернуть детальную информацию о товарах, но для простоты отдаем только список ID
	c.JSON(http.StatusOK, carts[userEmail])
}

// POST /api/checkout
func (oh *OrderHandler) Checkout(c *gin.Context) {
	userEmail, err := c.Cookie("session_token")
	if err != nil {
		c.String(http.StatusUnauthorized, "Пользователь не аутентифицирован")
		return
	}
	var payload struct {
		Email string `json:"email"`
	}
	if err := c.BindJSON(&payload); err != nil {
		c.String(http.StatusBadRequest, "Неверные данные")
		return
	}
	userCart := carts[userEmail]
	if len(userCart) == 0 {
		c.String(http.StatusBadRequest, "Корзина пуста")
		return
	}
	if err := oh.orderService.Checkout(userEmail, payload.Email, userCart); err != nil {
		c.String(http.StatusInternalServerError, "Ошибка оформления заказа")
		return
	}
	// Очищаем корзину после оформления заказа
	carts[userEmail] = []string{}
	c.String(http.StatusOK, "Заказ успешно оформлен")
}
