package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"strings"

	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/codeuiprogramming/e-commerce/app/consts"
	"github.com/codeuiprogramming/e-commerce/app/helpers"
	"github.com/codeuiprogramming/e-commerce/app/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/shopspring/decimal"
	"github.com/unrolled/render"
)

type CheckoutRequest struct {
	Cart *models.Cart
	ShippingFee *ShippingFee
	ShippingAddress *ShippingAddress
}

type ShippingFee struct {
	Courier string
	PackageName string
	Fee float64
}

type ShippingAddress struct {
	FirstName string
	LastName string
	CityID string
	ProvinceID string
	Address1 string
	Address2 string
	Phone string
	Email string
	PostCode string
}

func (server *Server) Checkout(w http.ResponseWriter, r *http.Request) {
    if !IsLoggedIn(r) {
        SetFlash(w, r, "error", "Anda perlu login!")
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return 
    }

    user := server.CurrentUser(w, r)

	shippingCost, err := server.getSelectedShippingCost(w,r)
	if err != nil {
		SetFlash(w, r, "error", "Proses checkout gagal")
		http.Redirect(w, r, "/carts", http.StatusSeeOther)
		return 
	}

	cartID := GetShoppingCartID(w, r)
	cart, _ := GetShoppingCart(server.DB, cartID)

	checkoutRequest := &CheckoutRequest{
		Cart: cart,
		ShippingFee: &ShippingFee{
			Courier:     r.FormValue("courier"),
			PackageName: r.FormValue("shipping_fee"),
			Fee:         shippingCost,
		},
		ShippingAddress: &ShippingAddress{
			FirstName: r.FormValue("first_name"),
			LastName:  r.FormValue("last_name"),
			CityID:    r.FormValue("city_id"),
			ProvinceID:r.FormValue("province_id"),
			Address1:  r.FormValue("address1"),
			Address2:  r.FormValue("address2"),
			Phone:     r.FormValue("phone"),
			Email:     r.FormValue("email"),
			PostCode:  r.FormValue("post_code"),
		},
	}

	order, err := server.SaveOrder(user, checkoutRequest)
	if err != nil {
		SetFlash(w, r, "error", "Proses checkout gagal")
		http.Redirect(w, r, "/carts", http.StatusSeeOther)
		return
	}

	ClearCart(server.DB, cartID)
	SetFlash(w, r, "success", "Data order berhasil disimpan")
	http.Redirect(w, r, "/orders/"+order.ID, http.StatusSeeOther)
}

func (server *Server) ShowOrder(w http.ResponseWriter, r *http.Request) {
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

	if vars["id"] == "" {
		http.Redirect(w, r, "/products", http.StatusSeeOther)
		return
	}

	orderModel := models.Order{}
	order, err := orderModel.FindByID(server.DB, vars["id"])
	if err != nil {
		http.Redirect(w, r, "/products", http.StatusSeeOther)
		return
	}

	if err := render.HTML(w, http.StatusOK, "show_order", server.DefaultRenderData(w, r, map[string]interface{}{
		"order":   order,
		"success": GetFlash(w, r, "success"),
		"snapToken":  order.PaymentToken.String, // token dari Midtrans
	})); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (server *Server) getSelectedShippingCost(w http.ResponseWriter, r *http.Request) (float64, error) {
    origin := os.Getenv("API_ONGKIR_ORIGIN")
    destination := r.FormValue("city_id")
    courier := r.FormValue("courier")
    shippingFeeSelected := r.FormValue("shipping_fee")

    cartID := GetShoppingCartID(w, r)
    cart, _ := GetShoppingCart(server.DB, cartID)

    if destination == "" {
        return 0, errors.New("invalid destination")
    }

    shippingFeeOptions, err := server.CalculateShippingFee(models.ShippingFeeParams{
        Origin:      origin,
        Destination: destination,
        Weight:      cart.TotalWeight,
        Courier:     courier,
    })
    if err != nil {
        return 0, errors.New("failed shipping calculation")
    }

    var shippingCost float64

    fmt.Println(">> Dari Form:", shippingFeeSelected)

    for _, option := range shippingFeeOptions {
        // Gabungkan beberapa info untuk memastikan cocok
		fullServiceName := fmt.Sprintf("%s - %d", option.Service, option.Fee)

        // Jika value dari form mengandung nama service atau sama persis dengan format kita
        if strings.Contains(shippingFeeSelected, option.Service) ||
           strings.Contains(shippingFeeSelected, fullServiceName) {

            shippingCost = float64(option.Fee)
            fmt.Println(">> Match ditemukan:", fullServiceName, "=>", shippingCost)
            break
        }
    }

    fmt.Println(">> Ongkir Terpilih Akhir:", shippingCost)
    return shippingCost, nil
}

func (server *Server) SaveOrder(user *models.User, r *CheckoutRequest) (*models.Order, error) {
	fmt.Println(">> Ongkir diterima di SaveOrder:", r.ShippingFee.Fee)
	var orderItems []models.OrderItem

	orderID := uuid.New().String()

	paymentURL, err := server.createdPaymentURL(user, r, orderID)
	if err != nil {
		return nil, err
	}

	if len(r.Cart.CartItems) > 0 {
		for _, cartItem := range r.Cart.CartItems {
			orderItems = append(orderItems, models.OrderItem{
				ProductID:       cartItem.ProductID,
				Qty:             cartItem.Qty,
				BasePrice:       cartItem.BasePrice,
				BaseTotal:       cartItem.BaseTotal,
				TaxAmount:       cartItem.TaxAmount,
				TaxPercent:      cartItem.TaxPercent,
				DiscountAmount:  cartItem.DiscountAmount,
				DiscountPercent: cartItem.DiscountPercent,
				SubTotal:        cartItem.SubTotal,
				Sku: func() string {
					if cartItem.Sku != "" { return cartItem.Sku }
					return cartItem.Product.Sku
				}(),
				Name: func() string {
					if cartItem.Name != "" { return cartItem.Name }
					return cartItem.Product.Name
				}(),
				Weight: cartItem.Product.Weight,
				DesignPath: cartItem.DesignPath,
				CustomType: cartItem.CustomType,
				CustomSize: cartItem.CustomSize,
			})
		}
	}

	orderCustomer := &models.OrderCustomer{
		UserID:     user.ID,
		FirstName:  r.ShippingAddress.FirstName,
		LastName:   r.ShippingAddress.LastName,
		CityID:     r.ShippingAddress.CityID,
		ProvinceID: r.ShippingAddress.ProvinceID,
		Address1:   r.ShippingAddress.Address1,
		Address2:   r.ShippingAddress.Address2,
		Phone:      r.ShippingAddress.Phone,
		Email:      r.ShippingAddress.Email,
		PostCode:   r.ShippingAddress.PostCode,
	}

	orderData := &models.Order{
    ID:                  orderID,
    UserID:              user.ID,
    OrderItems:          orderItems,
    OrderCustomer:       orderCustomer,
    Status:              0,
    OrderDate:           time.Now(),
    PaymentDue:          time.Now().AddDate(0, 0, 7),
    PaymentStatus:       consts.OrderPaymentStatusUnpaid,
    BaseTotalPrice:      r.Cart.BaseTotalPrice,
    TaxAmount:           r.Cart.TaxAmount,
    TaxPercent:          r.Cart.TaxPercent,
    DiscountAmount:      r.Cart.DiscountAmount,
    DiscountPercent:     r.Cart.DiscountPercent,
    ShippingCost:        decimal.NewFromFloat(r.ShippingFee.Fee),
    ShippingCourier:     r.ShippingFee.Courier,
    ShippingServiceName: r.ShippingFee.PackageName,
    PaymentToken:        sql.NullString{String: paymentURL, Valid: true},
	}

	orderData.GrandTotal = orderData.BaseTotalPrice.
    Add(orderData.TaxAmount).
    Sub(orderData.DiscountAmount).
    Add(orderData.ShippingCost)

	orderModel := models.Order{}
	order, err := orderModel.CreateOrder(server.DB, orderData)
	if err != nil {
		return nil, err
	}

	// Move design files into order-specific folder and update OrderItem.DesignPath
	for _, oi := range order.OrderItems {
		if oi.DesignPath == "" {
			continue
		}
		src := filepath.Join("public", oi.DesignPath)
		dir := filepath.Join("public", "uploads", "orders", order.ID)
		_ = os.MkdirAll(dir, 0755)
		filename := oi.ProductID + filepath.Ext(oi.DesignPath)
		dest := filepath.Join(dir, filename)
		if err := os.Rename(src, dest); err != nil {
			log.Println("SaveOrder: failed move design file", src, "->", dest, "err:", err)
		} else {
			newPath := filepath.ToSlash(filepath.Join("uploads", "orders", order.ID, filename))
			// update DB record for this order item
			server.DB.Model(&models.OrderItem{}).Where("id = ?", oi.ID).Update("design_path", newPath)
			log.Println("SaveOrder: moved design file to", newPath)
		}
	}

	// cleanup temporary custom products: if a product slug starts with "custom-",
	// remove the product DB record (images on disk were moved/kept)
	for _, oi := range order.OrderItems {
		var prod models.Product
		if err := server.DB.Where("id = ?", oi.ProductID).First(&prod).Error; err == nil {
			if strings.HasPrefix(prod.Slug, "custom-") {
				// delete product images records
				server.DB.Where("product_id = ?", prod.ID).Delete(&models.ProductImage{})
				// delete product record
				server.DB.Delete(&prod)
				log.Println("SaveOrder: removed temporary product", prod.ID)
			}
		}
	}

	return order, nil
}

func (server *Server) createdPaymentURL(user *models.User, r *CheckoutRequest, orderID string) (string, error) {
	midtransServerKey := os.Getenv("API_MIDTRANS_SERVER_KEY")
	midtrans.ServerKey = midtransServerKey

	var enabledPaymentTypes []snap.SnapPaymentType
	enabledPaymentTypes = append(enabledPaymentTypes, snap.AllSnapPaymentType...)

	// Tambahkan ShippingCost ke total pembayaran
	totalWithShipping := r.Cart.GrandTotal.Add(decimal.NewFromFloat(r.ShippingFee.Fee))

	snapRequest := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: totalWithShipping.IntPart(),
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: user.FirstName,
			LName: user.LastName,
			Email: user.Email,
		},
		EnabledPayments: enabledPaymentTypes,
	}

	snapResponse, err := snap.CreateTransaction(snapRequest)
	if err != nil {
		return "", err
	}

	return snapResponse.Token, nil
}
