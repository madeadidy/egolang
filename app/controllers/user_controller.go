package controllers

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/codeuiprogramming/e-commerce/app/helpers"
	"github.com/codeuiprogramming/e-commerce/app/models"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/unrolled/render"
)

func (server *Server) Login(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
		Layout:     "layout",
		Extensions: []string{".html", ".tmpl"},
		Funcs: []template.FuncMap{
			{
				"FormatPrice": helpers.FormatPrice,
			},
		},
	})

	_ = render.HTML(w, http.StatusOK, "login", server.DefaultRenderData(w, r, map[string]interface{}{
		"error": GetFlash(w, r, "error"),
	}))
}

func (server *Server) DoLogin(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	userModel := models.User{}
	user, err := userModel.FindByEmail(server.DB, email)
	if err != nil {
		SetFlash(w, r, "error", "email or password invalid")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if !ComparePassword(password, user.Password) {
		SetFlash(w, r, "error", "email or password invalid")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	session, _ := store.Get(r, sessionUser)
	session.Values["id"] = user.ID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (server *Server) Register(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
	Layout: "layout",
	Extensions: []string{".html",".tmpl"},
	})

	_ = render.HTML(w, http.StatusOK, "register", server.DefaultRenderData(w, r, map[string]interface{}{
		"error": GetFlash(w, r, "error"),
	}))
}

func (server *Server) DoRegister(w http.ResponseWriter, r *http.Request) {
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if firstName == "" || lastName == "" || email == "" || password == "" {
		SetFlash(w, r, "error", "First name, last name, email and password are required!")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	userModel := models.User{}
	existUser, err := userModel.FindByEmail(server.DB, email)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if existUser != nil {
		SetFlash(w, r, "error", "Sorry, email already registered")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	hashedPassword, _ := MakePassword(password)
	params := &models.User{
		ID:            uuid.New().String(),
		FirstName:     firstName,
		LastName:      lastName,
		Email:         email,
		Password:      hashedPassword,
	}
	params.Role = "user"

	user, err := userModel.CreateUser(server.DB, params)
	if err != nil {
		SetFlash(w, r, "error", "Sorry, registration failed")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}
	
	session, _ := store.Get(r, sessionUser)
	session.Values["id"] = user.ID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (server *Server) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, sessionUser)

	// Clear session value and expire cookie so browser removes it
	session.Values["id"] = nil
	// Set MaxAge < 0 to delete cookie according to gorilla/sessions docs
	if session.Options == nil {
		session.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   false,
		}
	} else {
		session.Options.MaxAge = -1
		session.Options.Path = "/"
	}
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (server *Server) Profile(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
		Layout:     "layout",
		Extensions: []string{".html", ".tmpl"},
		Funcs: []template.FuncMap{
			{
				"FormatPrice": helpers.FormatPrice,
			},
		},
	})

	data := server.DefaultRenderData(w, r, map[string]interface{}{})
	// include flash messages
	data["success"] = GetFlash(w, r, "success")
	data["error"] = GetFlash(w, r, "error")

	// Prefer avatar stored in DB; fallback to filesystem check
	u := server.CurrentUser(w, r)
	if u != nil {
		var avatar models.UserAvatar
		if err := server.DB.Where("user_id = ?", u.ID).First(&avatar).Error; err == nil {
			data["avatar"] = avatar.Path
		} else {
			pattern := filepath.Join("assets", "uploads", "avatars", u.ID+".*")
			matches, _ := filepath.Glob(pattern)
			if len(matches) > 0 {
				filename := filepath.Base(matches[0])
				data["avatar"] = "/public/uploads/avatars/" + filename
			}
		}

		// prepare prefill values for gender and dob
		if u.Gender == "male" {
			data["genderMale"] = "selected"
		} else if u.Gender == "female" {
			data["genderFemale"] = "selected"
		}
		if u.Dob != nil {
			data["dobValue"] = u.Dob.Format("2006-01-02")
		}
	}

	// load provinces server-side so template can render options immediately
	if provinces, err := server.GetProvince(); err == nil {
		data["provinces"] = provinces
	} else {
		data["provinces"] = nil
	}

	// load addresses for this user so template can render saved addresses
	if u != nil {
		var addrs []models.Address
		if err := server.DB.Where("user_id = ?", u.ID).Order("created_at desc").Find(&addrs).Error; err == nil {
			data["addresses"] = addrs
		} else {
			data["addresses"] = nil
		}
	} else {
		data["addresses"] = nil
	}

	_ = render.HTML(w, http.StatusOK, "profile", data)
}

// UpdateProfile handles POST from profile edit form, including avatar upload.
func (server *Server) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if !IsLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user := server.CurrentUser(w, r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// limit parse size (10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		SetFlash(w, r, "error", "failed to parse form")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	first := r.FormValue("first_name")
	last := r.FormValue("last_name")

	updates := map[string]interface{}{}
	if first != "" {
		updates["first_name"] = first
	}
	if last != "" {
		updates["last_name"] = last
	}

	// always collect dob, phone and gender into updates if provided
	if dobStr := r.FormValue("dob"); dobStr != "" {
		if t, err := time.Parse("2006-01-02", dobStr); err == nil {
			updates["dob"] = &t
		}
	}
	if phone := r.FormValue("phone"); phone != "" {
		updates["phone"] = phone
	}
	if gender := r.FormValue("gender"); gender != "" {
		updates["gender"] = gender
	}

	// handle avatar upload (optional)
	file, handler, err := r.FormFile("avatar")
	if err == nil {
		defer file.Close()

		// ensure upload dir exists (place under assets so it's served at /public/...)
		uploadDir := "assets/uploads/avatars"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			SetFlash(w, r, "error", "failed to create upload directory")
			http.Redirect(w, r, "/profile", http.StatusSeeOther)
			return
		}

		// remove any existing files for this user (cleanup old uploads)
		oldPattern := filepath.Join(uploadDir, user.ID+".*")
		oldFiles, _ := filepath.Glob(oldPattern)
		for _, f := range oldFiles {
			_ = os.Remove(f)
		}

		// also remove DB records for avatars for this user (ensure table exists)
		_ = server.DB.AutoMigrate(&models.UserAvatar{})
		_ = server.DB.Where("user_id = ?", user.ID).Delete(&models.UserAvatar{}).Error

		ext := filepath.Ext(handler.Filename)
		if ext == "" {
			ext = ".jpg"
		}
		filename := fmt.Sprintf("%s%s", user.ID, ext)
		dstPath := filepath.Join(uploadDir, filename)

		dst, err := os.Create(dstPath)
		if err != nil {
			SetFlash(w, r, "error", "failed to save uploaded file")
			http.Redirect(w, r, "/profile", http.StatusSeeOther)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			SetFlash(w, r, "error", "failed to write uploaded file")
			http.Redirect(w, r, "/profile", http.StatusSeeOther)
			return
		}

		// create DB record for this avatar (path uses public mapping)
		avatarPath := "/public/uploads/avatars/" + filename
		avatarRecord := &models.UserAvatar{
			ID:        uuid.New().String(),
			UserID:    user.ID,
			Path:      avatarPath,
			IsPrimary: true,
		}
		if err := server.DB.Create(avatarRecord).Error; err != nil {
			// if DB insert failed, remove the saved file and report error
			_ = os.Remove(dstPath)
			SetFlash(w, r, "error", "failed to save avatar metadata")
			http.Redirect(w, r, "/profile", http.StatusSeeOther)
			return
		}
	}

	if len(updates) > 0 {
		if err := server.DB.Model(&models.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
			SetFlash(w, r, "error", "failed to update profile")
			http.Redirect(w, r, "/profile", http.StatusSeeOther)
			return
		}
	}

	SetFlash(w, r, "success", "Profile updated successfully")
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

// CreateAddress handles creating a new address for the current user
func (server *Server) CreateAddress(w http.ResponseWriter, r *http.Request) {
	if !IsLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user := server.CurrentUser(w, r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		SetFlash(w, r, "error", "failed to parse form")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	phone := r.FormValue("phone")
	provinceID := r.FormValue("province_id")
	cityID := r.FormValue("city_id")
	// district field (stored) and postcode
	districtID := r.FormValue("district_id")
	postcode := strings.TrimSpace(r.FormValue("postcode"))
	address1 := r.FormValue("address1")
	address2 := r.FormValue("address2")

	addr := &models.Address{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Name:      name,
		IsPrimary: false,
		CityID:    cityID,
		ProvinceID: provinceID,
		DistrictID: districtID,
		Address1:  address1,
		Address2:  address2,
		Phone:     phone,
		PostCode:  postcode,
	}

	// simple postcode validation: must be 5 digits
	if postcode == "" {
		SetFlash(w, r, "error", "Kode pos harus diisi")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}
	m := regexp.MustCompile(`^[0-9]{5}$`)
	if !m.MatchString(postcode) {
		SetFlash(w, r, "error", "Kode pos tidak valid — harus 5 digit angka")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	if err := server.DB.Create(addr).Error; err != nil {
		SetFlash(w, r, "error", "failed to save address")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	fmt.Println("CreateAddress: saved address id=", addr.ID, "user=", user.ID, "city=", cityID, "district=", districtID, "postcode=", postcode)

	SetFlash(w, r, "success", "Alamat berhasil ditambahkan")
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

// DeleteAddress deletes an address owned by the current user
func (server *Server) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	if !IsLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user := server.CurrentUser(w, r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		SetFlash(w, r, "error", "failed to parse form")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		SetFlash(w, r, "error", "invalid address id")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	// ensure only user's own address is deleted
	if err := server.DB.Where("id = ? AND user_id = ?", id, user.ID).Delete(&models.Address{}).Error; err != nil {
		SetFlash(w, r, "error", "failed to delete address")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	SetFlash(w, r, "success", "Alamat berhasil dihapus")
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

// SetPrimaryAddress marks an address as the primary address for the current user
func (server *Server) SetPrimaryAddress(w http.ResponseWriter, r *http.Request) {
	if !IsLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user := server.CurrentUser(w, r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		SetFlash(w, r, "error", "failed to parse form")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		SetFlash(w, r, "error", "invalid address id")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	tx := server.DB.Begin()
	if tx.Error != nil {
		SetFlash(w, r, "error", "database error")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	// unset primary flag for all user's addresses
	if err := tx.Model(&models.Address{}).Where("user_id = ?", user.ID).Update("is_primary", false).Error; err != nil {
		tx.Rollback()
		SetFlash(w, r, "error", "failed to update addresses")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	// set primary on chosen address (safeguard by user_id)
	if err := tx.Model(&models.Address{}).Where("id = ? AND user_id = ?", id, user.ID).Update("is_primary", true).Error; err != nil {
		tx.Rollback()
		SetFlash(w, r, "error", "failed to set primary address")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		SetFlash(w, r, "error", "database error")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	SetFlash(w, r, "success", "Berhasil mengatur alamat sebagai utama")
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

// AdminLogin shows a dedicated admin login page (separate from public user login)
func (server *Server) AdminLogin(w http.ResponseWriter, r *http.Request) {
	render := newAdminRender()

	data := server.DefaultRenderData(w, r, map[string]interface{}{
		"error": GetFlash(w, r, "error"),
	})

	_ = render.HTML(w, http.StatusOK, "admin_login", data)
}

// DoAdminLogin handles admin authentication and ensures user role is admin
func (server *Server) DoAdminLogin(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	userModel := models.User{}
	user, err := userModel.FindByEmail(server.DB, email)
	if err != nil || user == nil {
		SetFlash(w, r, "error", "email or password invalid")
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	if !ComparePassword(password, user.Password) {
		SetFlash(w, r, "error", "email or password invalid")
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	if !(user.Role == "admin" || user.Role == "superadmin") {
		SetFlash(w, r, "error", "account does not have admin access")
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	session, _ := store.Get(r, sessionUser)
	session.Values["id"] = user.ID
	session.Save(r, w)

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}