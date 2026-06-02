package controllers

import (
	"net/http"
	"os"
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

	// Admin login (separate form for admin users)
	server.Router.HandleFunc("/admin/login", server.AdminLogin).Methods("GET")
	server.Router.HandleFunc("/admin/login", server.DoAdminLogin).Methods("POST")
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
	server.Router.HandleFunc("/carts/districts", server.GetDistrictsByCity).Methods("GET")
	server.Router.HandleFunc("/carts/postcodes", server.GetPostcodesByCity).Methods("GET")
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
	server.Router.HandleFunc("/profile/address/delete", server.DeleteAddress).Methods("POST")
	server.Router.HandleFunc("/profile/address/set-primary", server.SetPrimaryAddress).Methods("POST")

	// Admin area (served by admin2 SPA at /admin)
	// Products CRUD
	// API for product CRUD (JSON) - protected by RequireAdminAuth
	server.Router.Handle("/api/admin/products", server.RequireAdminAuth(http.HandlerFunc(server.APIAdminProducts))).Methods("GET", "POST")
	server.Router.Handle("/api/admin/products/{id}", server.RequireAdminAuth(http.HandlerFunc(server.APIAdminProduct))).Methods("GET", "PUT", "DELETE")
	// upload product images
	server.Router.Handle("/api/admin/products/{id}/images", server.RequireAdminAuth(http.HandlerFunc(server.APIAdminProductImageUpload))).Methods("POST")

	// API for user management (admin only)
	server.Router.Handle("/api/admin/users", server.RequireAdminAuth(http.HandlerFunc(server.APIAdminUsers))).Methods("GET")
	server.Router.Handle("/api/admin/users/{id}", server.RequireAdminAuth(http.HandlerFunc(server.APIAdminUser))).Methods("GET", "PUT", "DELETE")

    // API for orders (admin only)
    server.Router.Handle("/api/admin/orders", server.RequireAdminAuth(http.HandlerFunc(server.APIAdminOrders))).Methods("GET")
    server.Router.Handle("/api/admin/orders/{id}", server.RequireAdminAuth(http.HandlerFunc(server.APIAdminOrder))).Methods("GET")

	server.Router.HandleFunc("/material-dashboard-shadcn-vue", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/material-dashboard-shadcn-vue/dashboard", http.StatusMovedPermanently)
	}).Methods("GET")
	server.Router.Handle("/admin/products", server.RequireAdminAuth(http.HandlerFunc(server.AdminProducts))).Methods("GET")
	server.Router.Handle("/admin/products/new", server.RequireAdminAuth(http.HandlerFunc(server.AdminProductNew))).Methods("GET")
	server.Router.Handle("/admin/products", server.RequireAdminAuth(http.HandlerFunc(server.AdminProductCreate))).Methods("POST")
	server.Router.Handle("/admin/products/{id}/edit", server.RequireAdminAuth(http.HandlerFunc(server.AdminProductEdit))).Methods("GET")
	server.Router.Handle("/admin/products/{id}", server.RequireAdminAuth(http.HandlerFunc(server.AdminProductUpdate))).Methods("POST")
	server.Router.Handle("/admin/products/{id}/delete", server.RequireAdminAuth(http.HandlerFunc(server.AdminProductDelete))).Methods("POST")

	server.Router.HandleFunc("/material-dashboard-shadcn-vue/admin", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/material-dashboard-shadcn-vue/dashboard", http.StatusMovedPermanently)
	}).Methods("GET", "HEAD")
	// Orders and customers
	server.Router.Handle("/admin/orders", server.RequireAdminAuth(http.HandlerFunc(server.AdminOrders))).Methods("GET")
	server.Router.Handle("/admin/orders/{id}", server.RequireAdminAuth(http.HandlerFunc(server.AdminOrderDetail))).Methods("GET")
	server.Router.Handle("/admin/customers", server.RequireAdminAuth(http.HandlerFunc(server.AdminCustomers))).Methods("GET")
	server.Router.Handle("/admin/customers/{id}", server.RequireAdminAuth(http.HandlerFunc(server.AdminCustomerDetail))).Methods("GET")
	
	// Also serve assets under the SPA's build base path so absolute URLs resolve.
	// If a requested file exists in public/admin2 serve it, otherwise return index.html
	// so the SPA router can handle client-side routes like /admin.
	// Redirect legacy base paths to /admin/ to ensure correct asset base
	server.Router.HandleFunc("/material-dashboard-shadcn-vue", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin/", http.StatusMovedPermanently)
	}).Methods("GET")
	server.Router.HandleFunc("/material-dashboard-shadcn-vue/admin", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin/", http.StatusMovedPermanently)
	}).Methods("GET", "HEAD")

	// Legacy product route: serve SPA index so client-side router shows /products
	server.Router.HandleFunc("/material-dashboard-shadcn-vue/products", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		http.ServeFile(w, r, "public/admin2/index.html")
	}).Methods("GET", "HEAD")

	server.Router.PathPrefix("/material-dashboard-shadcn-vue/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/material-dashboard-shadcn-vue/")
		// protect against directory traversal
		cleanPath := filepath.Clean(p)
		filePath := filepath.Join("public", "admin2", cleanPath)
		if cleanPath == "" || cleanPath == "." {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			http.ServeFile(w, r, "public/admin2/index.html")
			return
		}
		if _, err := os.Stat(filePath); err == nil {
			http.ServeFile(w, r, filePath)
			return
		}
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		http.ServeFile(w, r, "public/admin2/index.html")
	}).Methods("GET", "HEAD")

	// Serve assets for older build base used during development/deploy (/Git/admin/)
	server.Router.PathPrefix("/Git/admin/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/Git/admin/")
		cleanPath := filepath.Clean(p)
		filePath := filepath.Join("public", "admin2", cleanPath)
		if cleanPath == "" || cleanPath == "." {
			http.ServeFile(w, r, "public/admin2/index.html")
			return
		}
		if _, err := os.Stat(filePath); err == nil {
			http.ServeFile(w, r, filePath)
			return
		}
		http.ServeFile(w, r, "public/admin2/index.html")
	}).Methods("GET", "HEAD")

	// Reports
	server.Router.HandleFunc("/admin/reports/monthly", server.AdminReportMonthly).Methods("GET")
	server.Router.HandleFunc("/admin/reports/monthly.csv", server.AdminReportCSV).Methods("GET")
	// API for reports (transactions JSON) used by admin SPA
	server.Router.Handle("/api/admin/reports/transactions", server.RequireAdminAuth(http.HandlerFunc(server.APIAdminReportTransactions))).Methods("GET")

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

	server.Router.PathPrefix("/public/").Handler(staticHandler).Methods("GET", "HEAD")

	// Serve uploaded files under /uploads/ -> ./public/uploads/
	server.Router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./public/uploads/")))).Methods("GET", "HEAD")

	// Serve built admin2 SPA (static files placed in public/admin2)
	// Serve index for /admin and static assets under /admin/
	server.Router.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/material-dashboard-shadcn-vue/dashboard", http.StatusMovedPermanently)
	}).Methods("GET", "HEAD")
	server.Router.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/admin2/index.html")
	}).Methods("GET", "HEAD")
	// Simple connect/test page to verify API connectivity
	server.Router.HandleFunc("/admin/connect", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/admin2/connect.html")
	}).Methods("GET", "HEAD")
	server.Router.PathPrefix("/admin/").Handler(http.StripPrefix("/admin/", http.FileServer(http.Dir("./public/admin2/")))).Methods("GET", "HEAD")
}