package controllers

import (
	"html/template"
	"log"
	"net/http"

	"github.com/codeuiprogramming/e-commerce/app/helpers"
	"github.com/codeuiprogramming/e-commerce/app/models"
	"github.com/shopspring/decimal"
	"github.com/unrolled/render"
)

func (server *Server) Home(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
		Layout:     "layout",
		Extensions: []string{".html", ".tmpl"},
		Funcs: []template.FuncMap{
			{
				"FormatPrice": helpers.FormatPrice,
			},
		},
	})

	// load a few products for trending section; if DB empty, use manual list
	productModel := models.Product{}
	productsPtr, _, err := productModel.GetProducts(server.DB, 4, 1)
	var products []models.Product
	if err != nil || productsPtr == nil || len(*productsPtr) == 0 {
		products = []models.Product{
			{
				Name:  "Kaos Artha",
				Slug:  "kaos-artha",
				Price: decimal.NewFromFloat(75000),
				Images: []models.ProductImage{{Path: "img/products/BajuArthaDepan.png"}},
			},
			{
				Name:  "Kaos Kita Semua Bersaudara",
				Slug:  "kaos-kita-semua-bersaudara",
				Price: decimal.NewFromFloat(85000),
				Images: []models.ProductImage{{Path: "img/products/BajuKSBDepan.png"}},
			},
			{
				Name:  "Kaos Nyepi Kertabuana",
				Slug:  "kaos-nyepi-kertabuana",
				Price: decimal.NewFromFloat(80000),
				Images: []models.ProductImage{{Path: "img/products/BajuNyepiL4.png"}},
			},
			{
				Name:  "Kaos Peradah Kaltim",
				Slug:  "kaos-peradah-kaltim",
				Price: decimal.NewFromFloat(125000),
				Images: []models.ProductImage{{Path: "img/products/BajuHitamPeradahKaltimDepan.png"}},
			},
		}
	} else {
		products = *productsPtr
	}

	log.Printf("Home: trending products count=%d", len(products))

	_ = render.HTML(w, http.StatusOK, "home", server.DefaultRenderData(w, r, map[string]interface{}{
		"products": products,
	}))
}