package controllers

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/codeuiprogramming/e-commerce/app/helpers"
	"github.com/codeuiprogramming/e-commerce/app/models"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

func (server *Server) Products(w http.ResponseWriter, r *http.Request) {
render := render.New(render.Options{
    Layout:     "layout",
    Extensions: []string{".html", ".tmpl"},
    Funcs: []template.FuncMap{
        {
            "FormatPrice": helpers.FormatPrice,
        },
    },
})


	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page <= 0 {
		page = 1
	}

	perPage := 9

	productModel := models.Product{}
	products, totalRows, err := productModel.GetProducts(server.DB, perPage, page)
	if err != nil {
		return
	}

	pagination, _ := GetPaginationLinks(server.AppConfig, PaginationParams{
		Path:        "products",
		TotalRows:   int32(totalRows),
		PerPage:     int32(perPage),
		CurrentPage: int32(page),
	})

	_ = render.HTML(w, http.StatusOK, "products", server.DefaultRenderData(w, r, map[string]interface{}{
		"products":   products,
		"pagination": pagination,
	}))
}

func (server *Server) GetProductBySlug(w http.ResponseWriter, r *http.Request) {
render := render.New(render.Options{
    Layout:     "layout",
    Extensions: []string{".html", ".tmpl"},
    Funcs: []template.FuncMap{
        {
            "FormatPrice": helpers.FormatPrice,
        },
    },
})


	vars := mux.Vars(r)

	if vars["slug"] == "" {
		return
	}

	productModel := models.Product{}
	product, err := productModel.FindBySlug(server.DB, vars["slug"])
	if err != nil {
		// product not found -> return 404 to client
		http.NotFound(w, r)
		return
	}

	_ = render.HTML(w, http.StatusOK, "product", server.DefaultRenderData(w, r, map[string]interface{}{
		"product": product,
		"success": GetFlash(w, r, "success"),
		"error":   GetFlash(w, r, "error"),
	}))
}

// PRODUCT CUSTOM
func (server *Server) ProductCustom(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
		Layout:     "layout",
		Extensions: []string{".html", ".tmpl"},
		Funcs: []template.FuncMap{
			{
				"FormatPrice": helpers.FormatPrice,
			},
		},
	})

	_ = render.HTML(w, http.StatusOK, "product-custom", server.DefaultRenderData(w, r, map[string]interface{}{}))
}

func (server *Server) ProductCustomList(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
		Layout:     "layout",
		Extensions: []string{".html", ".tmpl"},
		Funcs: []template.FuncMap{
			{
				"FormatPrice": helpers.FormatPrice,
			},
		},
	})

	customProducts := []map[string]string{
		{"Name": "Kaos", "Slug": "tshirt", "Image": "public/img/products/KaosCustom.jpg"},
		{"Name": "Mug", "Slug": "mug", "Image": "public/img/products/MugCustom.jpg"},
		{"Name": "Totebag", "Slug": "totebag", "Image": "public/img/products/TotebagCustom.jpg"},
	}

	_ = render.HTML(w, http.StatusOK, "product-custom-list", server.DefaultRenderData(w, r, map[string]interface{}{
		"customProducts": customProducts,
	}))
}

// func (server *Server) ProductCustomDetail(w http.ResponseWriter, r *http.Request) {
//     render := render.New(render.Options{
//         Layout:     "layout",
//         Extensions: []string{".html", ".tmpl"},
//         Funcs: []template.FuncMap{
//             {
//                 "FormatPrice": helpers.FormatPrice,
//             },
//         },
//     })

//     vars := mux.Vars(r)
//     productType := vars["type"]

//     // mapping gambar berdasarkan type
//     imageMap := map[string]string{
//         "tshirt":  "img/products/KaosCustom.jpg",
//         "mug":     "img/products/MugCustom.jpg",
//         "totebag": "img/products/TotebagCustom.jpg",
//     }

//     // default jika tidak ditemukan
//     baseImage := imageMap[productType]
//     if baseImage == "" {
//         baseImage = "img/products/default-custom.jpg"
//     }

//     _ = render.HTML(w, http.StatusOK, "product-custom-detail", map[string]interface{}{
//         "user":  server.CurrentUser(w, r),
//         "type":  productType,
//         "Image": baseImage,
//     })
// }

func (server *Server) ProductCustomDetail(w http.ResponseWriter, r *http.Request) {
    render := render.New(render.Options{
        Layout:     "layout",
        Extensions: []string{".html", ".tmpl"},
        Funcs: []template.FuncMap{
            {
                "FormatPrice": helpers.FormatPrice,
            },
        },
    })

    vars := mux.Vars(r)
    slug := vars["type"]

	// Ambil data produk berdasarkan slug
	var product models.CustomProduct
	if err := server.DB.Where("slug = ?", slug).First(&product).Error; err != nil {
		// Jika tidak ditemukan di DB, gunakan fallback manual per tipe supaya tidak 404
		imageMap := map[string]string{
			"tshirt":  "img/products/KaosCustom.jpg",
			"mug":     "img/products/MugCustom.jpg",
			"totebag": "img/products/TotebagCustom.jpg",
		}

		// default values per type
		switch slug {
		case "tshirt":
			product = models.CustomProduct{Type: "tshirt", Name: "Kaos", Slug: "tshirt", BasePrice: 75000, CustomFee: 15000, Description: "Custom kaos - cetak desain Anda"}
		case "mug":
			product = models.CustomProduct{Type: "mug", Name: "Mug", Slug: "mug", BasePrice: 40000, CustomFee: 10000, Description: "Custom mug - cetak desain Anda"}
		case "totebag":
			product = models.CustomProduct{Type: "totebag", Name: "Totebag", Slug: "totebag", BasePrice: 60000, CustomFee: 12000, Description: "Custom totebag - cetak desain Anda"}
		default:
			product = models.CustomProduct{Type: slug, Name: slug, Slug: slug, BasePrice: 50000, CustomFee: 10000, Description: "Custom product"}
		}

		baseImage := imageMap[slug]
		if baseImage == "" {
			baseImage = "img/products/default-custom.jpg"
		}

		_ = render.HTML(w, http.StatusOK, "product-custom-detail", server.DefaultRenderData(w, r, map[string]interface{}{
			"type":      product.Type,
			"Image":     baseImage,
			"BasePrice": product.BasePrice,
			"CustomFee": product.CustomFee,
			"Product":   product,
		}))
		return
	}

	// Ambil main image (gambar utama)
	var mainImage models.CustomProductImage
	server.DB.
		Where("custom_product_id = ? AND is_main = true", product.ID).
		First(&mainImage)

	// Fallback image bawaan per tipe
	imageMap := map[string]string{
		"tshirt":  "img/products/KaosCustom.jpg",
		"mug":     "img/products/MugCustom.jpg",
		"totebag": "img/products/TotebagCustom.jpg",
	}

	// Tentukan image yang dipakai
	baseImage := mainImage.ImageURL
	if baseImage == "" {
		baseImage = imageMap[product.Type]
	}
	if baseImage == "" {
		baseImage = "img/products/default-custom.jpg"
	}

	// Render HTML
	_ = render.HTML(w, http.StatusOK, "product-custom-detail", server.DefaultRenderData(w, r, map[string]interface{}{
		"type":      product.Type,
		"Image":     baseImage,
		"BasePrice": product.BasePrice,
		"CustomFee": product.CustomFee,
		"Product":   product,
	}))
}

