package handlers

import (
	"net/http"

	"github.com/LOOK-MOM-I-CAN-FLY/E-Commerce/internal/service"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(ps *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: ps}
}

// GET /api/products
func (ph *ProductHandler) GetProducts(c *gin.Context) {
	products := ph.productService.GetProducts()
	c.JSON(http.StatusOK, products)
}
