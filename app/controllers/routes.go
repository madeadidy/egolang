package controllers

import (
	"net/http"
	"path/filepath"
	"strings"

	// "github.com/codeuiprogramming/e-commerce/app/controllers"
	"github.com/gorilla/mux"
)

func (server *Server) initializeRoutes() {
	server.Router = mux.NewRouter()
	// Middleware global: baca header Authorization Bearer dan simpan token di context
	server.Router.Use(server.AuthHeaderMiddleware)
	server.Router.HandleFunc("/", server.Home).Methods("GET")
	
	server.Router.HandleFunc("/login", server.Login).Methods("GET")
	server.Router.HandleFunc("/login", server.DoLogin).Methods("POST")
	server.Router.HandleFunc("/register", server.Register).Methods("GET")
	server.Router.HandleFunc("/register", server.DoRegister).Methods("POST")
	server.Router.HandleFunc("/logout", server.Logout).Methods("GET")

	server.Router.HandleFunc("/products", server.Products).Methods("GET")
	server.Router.HandleFunc("/products/{slug}", server.GetProductBySlug).Methods("GET")

	// server.Router.HandleFunc("/product-custom", server.ProductCustom).Methods("GET")
	// server.Router.HandleFunc("/product-custom", server.ProductCustomList).Methods("GET")
	server.Router.HandleFunc("/product-custom", server.ProductCustomList).Methods("GET")
	server.Router.HandleFunc("/product-custom/{type}", server.ProductCustomDetail).Methods("GET")

	server.Router.HandleFunc("/carts", server.GetCart).Methods("GET")
	server.Router.HandleFunc("/carts", server.AddItemToCart).Methods("POST")
	server.Router.HandleFunc("/carts/custom", server.AddCustomToCart).Methods("POST")
	server.Router.HandleFunc("/carts/update", server.UpdateCart).Methods("POST")
	server.Router.HandleFunc("/carts/cities", server.GetCitiesByProvince).Methods("GET")
	server.Router.HandleFunc("/provinces", server.GetProvinces).Methods("GET")
	server.Router.HandleFunc("/carts/calculate-shipping", server.CalculateShipping).Methods("POST")
	server.Router.HandleFunc("/carts/apply-shipping", server.ApplyShipping).Methods("POST")
	server.Router.HandleFunc("/carts/remove/{id}", server.RemoveItemByID).Methods("GET")

	// Lindungi route checkout dengan middleware AuthRequired (menggunakan session lama).
	server.Router.Handle("/orders/checkout", server.AuthRequired(http.HandlerFunc(server.Checkout))).Methods("POST")
	server.Router.HandleFunc("/orders/{id}", server.ShowOrder).Methods("GET")
	// Profile page
	server.Router.HandleFunc("/profile", server.Profile).Methods("GET")
	server.Router.HandleFunc("/profile", server.UpdateProfile).Methods("POST")
	server.Router.HandleFunc("/profile/address", server.CreateAddress).Methods("POST")

	// Endpoint untuk menerima klaim session dari Clerk frontend (development flow)
	server.Router.HandleFunc("/auth/clerk/claim", server.ClaimClerk).Methods("POST")

	// Debug: lihat apakah server mengenali session saat ini
	server.Router.HandleFunc("/_debug/session", server.DebugSession).Methods("GET")

	// server.Router.HandleFunc("/payments/midtrans", server.Midtrans).Methods("POST")
	server.Router.HandleFunc("/payments/notification", server.PaymentNotification).Methods("POST")

	// Serve static files: prefer assets/ then fallback to public/
	assetsFS := http.Dir("./assets/")
	publicFS := http.Dir("./public/")

	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// trim prefix '/public/'
		p := strings.TrimPrefix(r.URL.Path, "/public/")
		// try assets first
		if f, err := assetsFS.Open(p); err == nil {
			f.Close()
			http.ServeFile(w, r, filepath.Join("assets", p))
			return
		}
		// fallback to public
		if f, err := publicFS.Open(p); err == nil {
			f.Close()
			http.ServeFile(w, r, filepath.Join("public", p))
			return
		}
		http.NotFound(w, r)
	})

	server.Router.PathPrefix("/public/").Handler(staticHandler).Methods("GET")
}