package models

import (
	"log"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Product struct {
	ID               string `gorm:"size:36;not null;uniqueIndex;primary_key"`
	ParentID         string `gorm:"size:36;index"`
	User             User
	UserID           string          `gorm:"size:36;index"`
	ProductImages    []ProductImage
	Categories       []Category 	 `gorm:"many2many:product_categories;"`
	Sku              string          `gorm:"size:100;index"`
	Name             string          `gorm:"size:255"`
	Slug             string          `gorm:"size:255"`
	Price            decimal.Decimal `gorm:"type:decimal(16,2);"`
	Stock            int
	Weight           decimal.Decimal `gorm:"type:decimal(10,2);"`
	ShortDescription string          `gorm:"type:text"`
	Description      string          `gorm:"type:text"`
	Status           int             `gorm:"default:0"`
	// IsTemporary menandai produk yang dibuat sementara untuk custom items
	IsTemporary      bool            `gorm:"default:false"`
	CreatedAt        time.Time
	UpdateAt         time.Time
	DeleteAt         gorm.DeletedAt
 	Images      	 []ProductImage `gorm:"foreignKey:ProductID"`
}

func (p *Product) GetProducts(db *gorm.DB, perPage int, page int) (*[]Product, int64, error) {
	var err error
	var products []Product
	var count int64

	// only count non-temporary products for public listing
	err = db.Model(&Product{}).Where("is_temporary = ?", false).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage

	err = db.
		Where("is_temporary = ?", false).
		Preload("ProductImages").
		Order(`
		CASE WHEN name = 'Totebag Barong' THEN 9999 ELSE 1 END,
		created_at DESC
	`).
		Limit(perPage).
		Offset(offset).
		Find(&products).Error

		for _, p := range products {
    log.Println("Loaded product:", p.Name, "Price:", p.Price.StringFixed(2))
}

	if err != nil {
		return nil, 0, err
	}

	return &products, count, nil
}

func (p *Product) FindBySlug(db *gorm.DB, slug string) (*Product, error) {
	var err error
	var product Product

	err = db.Debug().Preload("ProductImages").Model(&Product{}).Where("slug = ?", slug).First(&product).Error
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (p *Product) FindByID(db *gorm.DB, id string) (Product, error) {
    var product Product
    if err := db.Where("id = ?", id).First(&product).Error; err != nil {
        return product, err
    }
    return product, nil
}
