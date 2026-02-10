package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/codeuiprogramming/e-commerce/app/helpers"
	"github.com/codeuiprogramming/e-commerce/app/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"

	"github.com/unrolled/render"
	"gorm.io/gorm"
)
func GetShoppingCartID(w http.ResponseWriter, r *http.Request) string {
	session, _ := store.Get(r, sessionShoppingCart) //  "shopping-cart-session" sessionShoppingCart
	if session.Values["cart-id"] == nil {
		session.Values["cart-id"] = uuid.New().String()
		session.Save(r, w)
		fmt.Println("cart-id in session:", session.Values["cart-id"])
	}
	return fmt.Sprintf("%v", session.Values["cart-id"])
}

func ClearCart(db *gorm.DB, cartID string) error {
    var cart models.Cart

    err := cart.ClearCart(db, cartID)
    if err != nil {
        return err
    }

    return nil
}

func GetShoppingCart(db *gorm.DB, cartID string) (*models.Cart, error) {
	var cart models.Cart

	existCart, err := cart.GetCart(db, cartID)
	if err != nil {
		existCart, _ = cart.CreateCart(db, cartID)
	}

	// Panggil CalculateCart dengan shippingFee (misal 0 dulu)
	_, _ = existCart.CalculateCart(db, cartID, decimal.NewFromInt(0))

	updatedCart, _ := cart.GetCart(db, cartID)

	totalWeight := 0
	productModel := models.Product{}
	for _, cartItem := range updatedCart.CartItems {
		product, _ := productModel.FindByID(db, cartItem.ProductID)

		productWeight, _ := product.Weight.Float64()
		ceilWeight := math.Ceil(productWeight)

		itemWeight := cartItem.Qty * int(ceilWeight)

		totalWeight += itemWeight
	}

	updatedCart.TotalWeight = totalWeight

	return updatedCart, nil
}


func (server *Server) GetCart(w http.ResponseWriter, r *http.Request) {
    render := render.New(render.Options{
        Layout:     "layout",
        Extensions: []string{".html", ".tmpl"},
        Funcs: []template.FuncMap{
            {
                "FormatPrice": helpers.FormatPrice,
            },
        },
    })

    var cart *models.Cart

    cartID := GetShoppingCartID(w, r)
    cart, _ = GetShoppingCart(server.DB, cartID)

    provinces, err := server.GetProvince()
    if err != nil {
        log.Fatal(err)
    }

    _ = render.HTML(w, http.StatusOK, "cart", server.DefaultRenderData(w, r, map[string]interface{}{
        "cart":      cart,
        "items":     cart.CartItems,
        "provinces": provinces,
        "success":   GetFlash(w, r, "success"),
        "error":     GetFlash(w, r, "error"),
    }))
}

func (server *Server) AddItemToCart(w http.ResponseWriter, r *http.Request) {
    productID := r.FormValue("product_id")
    qty, _ := strconv.Atoi(r.FormValue("qty"))

    productModel := models.Product{}
    product, err := productModel.FindByID(server.DB, productID)
    if err != nil {
        SetFlash(w, r, "error", "Produk tidak ditemukan")
        http.Redirect(w, r, "/products", http.StatusSeeOther)
        return
    }

    if qty > product.Stock {
        SetFlash(w, r, "error", "Stok tidak mencukupi")
        http.Redirect(w, r, "/products/"+product.Slug, http.StatusSeeOther)
        return
    }

    cartID := GetShoppingCartID(w, r)
    cart, _ := GetShoppingCart(server.DB, cartID)

    _, err = cart.AddItem(server.DB, models.CartItem{
        ProductID: productID,
        Qty:       qty,
    })
    if err != nil {
        SetFlash(w, r, "error", "Gagal menambahkan ke keranjang")
        http.Redirect(w, r, "/products/"+product.Slug, http.StatusSeeOther)
        return
    }

    SetFlash(w, r, "success", "Item berhasil ditambahkan")
    http.Redirect(w, r, "/carts", http.StatusSeeOther)
}


func (server *Server) UpdateCart(w http.ResponseWriter, r *http.Request) {
	cartID := GetShoppingCartID(w, r)
	cart, _ := GetShoppingCart(server.DB, cartID)

	for _, item := range cart.CartItems {
		qty, _ := strconv.Atoi(r.FormValue(item.ID))

		_, err := cart.UpdateItemQty(server.DB, item.ID, qty)
		if err != nil {
			http.Redirect(w, r, "/carts", http.StatusSeeOther)
		}
	}
	http.Redirect(w, r, "/carts", http.StatusSeeOther)
}

func (server *Server) RemoveItemByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if vars["id"] == "" {
		http.Redirect(w, r, "/carts", http.StatusSeeOther)
	}
	
	cartID := GetShoppingCartID(w, r)
	cart, _ := GetShoppingCart(server.DB, cartID)

	err := cart.RemoveItemByID(server.DB, vars["id"])
	if err != nil {
		http.Redirect(w, r, "/carts", http.StatusSeeOther)
	}
	
	http.Redirect(w, r, "/carts", http.StatusSeeOther)
}

func (server *Server) GetCitiesByProvince(w http.ResponseWriter, r *http.Request) {
	provinceID := r.URL.Query().Get("province_id")
	if provinceID == "" {
    http.Error(w, "province_id is required", http.StatusBadRequest)
    return
	}

	cities, err := server.GetCitiesByProvinceID(provinceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
    	return
	}

	res := Result{Code: 200, Data: cities, Message: "Success"}
	result, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

// GetProvinces returns list of provinces (proxy to external API wrapper)
func (server *Server) GetProvinces(w http.ResponseWriter, r *http.Request) {
    provinces, err := server.GetProvince()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    res := Result{Code: 200, Data: provinces, Message: "Success"}
    result, err := json.Marshal(res)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(result)
}

type ShippingRequest struct {
    CityID  string `json:"city_id"`
    Courier string `json:"courier"`
}

func (server *Server) CalculateShipping(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    log.Println("RAW BODY:", string(body)) // debug

    var req ShippingRequest
    if err := json.Unmarshal(body, &req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    origin := os.Getenv("API_ONGKIR_ORIGIN")
    destination := req.CityID
    courier := req.Courier

    if destination == "" {
        http.Error(w, "invalid destination", http.StatusBadRequest)
        return
    }

    cartID := GetShoppingCartID(w, r)
    cart, _ := GetShoppingCart(server.DB, cartID)

    shippingFeeOptions, err := server.CalculateShippingFee(models.ShippingFeeParams{
        Origin:      origin,
        Destination: destination,
        Weight:      cart.TotalWeight,
        Courier:     courier,
    })

    if err != nil {
        http.Error(w, "calculation failed", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "meta": map[string]interface{}{
            "message": "Success calculate shipping",
            "code":    200,
            "status":  "success",
        },
        "data": shippingFeeOptions,
    })
}

type ApplyShippingRequest struct {
    ShippingPackage string `json:"shipping_package"`
    CityID          string `json:"city_id"`
    Courier         string `json:"courier"`
}

func (server *Server) ApplyShipping(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    var req ApplyShippingRequest
    if err := json.Unmarshal(body, &req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    origin := os.Getenv("API_ONGKIR_ORIGIN")
    destination := req.CityID
    courier := req.Courier
    shippingPackage := req.ShippingPackage

    cartID := GetShoppingCartID(w, r)
    cart, _ := GetShoppingCart(server.DB, cartID)

    if destination == "" {
        http.Error(w, "invalid destination", http.StatusInternalServerError)
        return
    }

    shippingFeeOptions, err := server.CalculateShippingFee(models.ShippingFeeParams{
        Origin:      origin,
        Destination: destination,
        Weight:      cart.TotalWeight,
        Courier:     courier,
    })
    if err != nil {
        http.Error(w, "invalid shipping calculation", http.StatusInternalServerError)
        return
    }

    var selectedShipping models.ShippingFeeOption
    for _, shippingOption := range shippingFeeOptions {
        if shippingOption.Service == shippingPackage {
            selectedShipping = shippingOption
            break
        }
    }

    cartGrandTotal, _ := cart.GrandTotal.Float64()
    grandTotal := cartGrandTotal + float64(selectedShipping.Fee)

    res := Result{Code: 200, Data: map[string]interface{}{
        "total_order":  cart.GrandTotal,
        "shipping_fee": selectedShipping.Fee,
        "grand_total":  grandTotal,
        "total_weight": cart.TotalWeight,
    }, Message: "Success"}

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(res)
}

// AddCustomToCart handles custom product submissions (design image + params)
func (server *Server) AddCustomToCart(w http.ResponseWriter, r *http.Request) {
    // read JSON body
    var payload struct {
        Type      string          `json:"type"`
        Size      string          `json:"size"`
        Quantity  int             `json:"quantity"`
        BasePrice json.RawMessage `json:"base_price"`
        CustomFee json.RawMessage `json:"custom_fee"`
        Design    string          `json:"design"`
    }

    body, err := io.ReadAll(r.Body)
    if err != nil {
        log.Println("AddCustomToCart: read body error:", err)
        http.Error(w, "failed to read body", http.StatusBadRequest)
        return
    }
    log.Println("AddCustomToCart: body len=", len(body))
    if len(body) > 0 && len(body) < 2000 {
        log.Println("AddCustomToCart body preview:", string(body))
    } else if len(body) >= 2000 {
        log.Println("AddCustomToCart body preview (first 2000 chars):", string(body[:2000]))
    }
    if err := json.Unmarshal(body, &payload); err != nil {
        log.Println("AddCustomToCart: json unmarshal error:", err)
        http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
        return
    }

    // parse BasePrice and CustomFee which may be numbers or formatted strings
    parseNumber := func(b json.RawMessage) float64 {
        if len(b) == 0 {
            return 0
        }
        // try parse as number
        var f float64
        if err := json.Unmarshal(b, &f); err == nil {
            return f
        }
        // try parse as string then clean
        var s string
        if err := json.Unmarshal(b, &s); err == nil {
            // remove non-digit, non-dot, non-comma, non-minus
            re := regexp.MustCompile(`[^0-9\.,\-]`)
            cleaned := re.ReplaceAllString(s, "")
            // remove thousand separators (dots), convert comma to dot
            cleaned = strings.ReplaceAll(cleaned, ".", "")
            cleaned = strings.ReplaceAll(cleaned, ",", ".")
            if cleaned == "" {
                return 0
            }
            if v, err := strconv.ParseFloat(cleaned, 64); err == nil {
                return v
            }
        }
        return 0
    }

    bp := parseNumber(payload.BasePrice)
    cf := parseNumber(payload.CustomFee)

    // create a temporary Product to represent this custom item
    priceTotal := bp + cf

    prod := models.Product{
        ID:    uuid.New().String(),
        Name:  "Custom - " + payload.Type,
        Slug:  "custom-" + uuid.New().String(),
        Price: decimal.NewFromFloat(priceTotal),
        Stock: 9999,
        Weight: decimal.NewFromFloat(0),
        IsTemporary: true,
    }

    // assign a valid UserID to satisfy foreign key constraint
    if u := server.CurrentUser(w, r); u != nil {
        prod.UserID = u.ID
    } else {
        // pick any existing user as owner (fallback)
        var anyUser models.User
        if err := server.DB.Limit(1).Find(&anyUser).Error; err == nil && anyUser.ID != "" {
            prod.UserID = anyUser.ID
        }
    }

    // save product
    if err := server.DB.Create(&prod).Error; err != nil {
        log.Println("AddCustomToCart: DB create product error:", err, "product:", prod)
        http.Error(w, "failed to create product: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // persist design image if present
    designPath := ""
    if strings.HasPrefix(payload.Design, "data:") {
        idx := strings.Index(payload.Design, ",")
        if idx > -1 {
            raw := payload.Design[idx+1:]
            data, err := base64.StdEncoding.DecodeString(raw)
            if err == nil {
                // ensure directory
                dir := "public/uploads/custom"
                _ = os.MkdirAll(dir, 0755)
                filename := prod.ID + ".png"
                path := filepath.Join(dir, filename)
                if err := ioutil.WriteFile(path, data, 0644); err == nil {
                    designPath = "uploads/custom/" + filename
                    // create product image record
                    img := models.ProductImage{ProductID: prod.ID, Path: designPath}
                    _ = server.DB.Create(&img)
                } else {
                    log.Println("AddCustomToCart: failed write design image:", err)
                }
            } else {
                log.Println("AddCustomToCart: decode design image error:", err)
            }
        }
    }

    // add to cart
    cartID := GetShoppingCartID(w, r)
    cart, _ := GetShoppingCart(server.DB, cartID)

    item := models.CartItem{
        ProductID:  prod.ID,
        Qty:        payload.Quantity,
        DesignPath: designPath,
        CustomType: payload.Type,
        CustomSize: payload.Size,
    }
    _, err = cart.AddItem(server.DB, item)
    if err != nil {
        http.Error(w, "failed to add to cart", http.StatusInternalServerError)
        return
    }

    // redirect to cart
    http.Redirect(w, r, "/carts", http.StatusSeeOther)
}
