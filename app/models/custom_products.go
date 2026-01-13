package models

import (
	"time"

	"gorm.io/gorm"
)

type CustomProduct struct {
	ID          int       `gorm:"column:id;primaryKey"`
	Name        string    `gorm:"column:name"`
	Slug        string    `gorm:"column:slug"`
	Type        string    `gorm:"column:type"`
	BasePrice   float64   `gorm:"column:base_price"`
	CustomFee   float64   `gorm:"column:custom_fee"`
	Description string    `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
	
	// Relasi dengan tabel gambar
	Images []CustomProductImage `gorm:"foreignKey:CustomProductID"`
}

func (cp *CustomProduct) GetAll(db *gorm.DB) ([]CustomProduct, error) {
	var products []CustomProduct
	err := db.Preload("Images", "is_main = true").Find(&products).Error
	return products, err
}
