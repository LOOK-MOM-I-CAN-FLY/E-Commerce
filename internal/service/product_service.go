package service

import (
	"github.com/LOOK-MOM-I-CAN-FLY/E-Commerce/internal/models"
	"github.com/LOOK-MOM-I-CAN-FLY/E-Commerce/internal/repository"
)

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (ps *ProductService) GetProducts() []models.Product {
	return ps.repo.GetAll()
}

func (ps *ProductService) GetProduct(id string) (models.Product, bool) {
	return ps.repo.GetByID(id)
}
