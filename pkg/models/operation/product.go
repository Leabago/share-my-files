package operation

import (
	"share-my-file/pkg/models"

	"gorm.io/gorm"
)

type product struct {
	DB *gorm.DB
}

func (product *product) Insert() {
	product.DB.Create(&models.Product{Code: "D42", Price: 100})
}
