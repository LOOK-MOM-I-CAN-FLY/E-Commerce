package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/LOOK-MOM-I-CAN-FLY/E-Commerce/internal/models"
)

type OrderRepository struct {
	mu     sync.Mutex
	orders map[string]models.Order
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		orders: make(map[string]models.Order),
	}
}

func (r *OrderRepository) Create(order models.Order) {
	r.mu.Lock()
	defer r.mu.Unlock()
	order.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	r.orders[order.ID] = order
}

// Дополнительные методы для получения заказов можно добавить при необходимости
