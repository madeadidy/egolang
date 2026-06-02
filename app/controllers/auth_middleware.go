package controllers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type ctxKey string

const (
    ctxKeyAuthToken ctxKey = "auth-token"
)

// AuthRequired middleware memastikan pengguna sudah login (menggunakan session lama).
// Ini adalah stub yang bekerja dengan session cookie yang sudah ada di project.
// Nantinya middleware ini bisa disesuaikan untuk memverifikasi Clerk session/token.
func (server *Server) AuthRequired(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !IsLoggedIn(r) {
            // Jika belum login, redirect ke halaman login (frontend Clerk akan menangani flow jika tersedia)
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        // Lanjutkan
        next.ServeHTTP(w, r)
    })
}

// AuthHeaderMiddleware membaca header `Authorization: Bearer <token>` dan menyimpan token
// ke dalam context request agar handler lain dapat mengaksesnya. Middleware ini TIDAK
// melakukan verifikasi token; ini hanya placeholder untuk integrasi server-side nanti.
func (server *Server) AuthHeaderMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        auth := r.Header.Get("Authorization")
        if auth != "" {
            // ekspektasi: "Bearer <token>"
            parts := strings.SplitN(auth, " ", 2)
            if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
                token := parts[1]
                // log incoming auth header for debugging
                fmt.Println("AuthHeaderMiddleware: Authorization header received (masked): Bearer ****")
                ctx := context.WithValue(r.Context(), ctxKeyAuthToken, token)
                r = r.WithContext(ctx)
            }
        }

        next.ServeHTTP(w, r)
    })
}

// RequireAdminAuth mengizinkan akses jika salah satu terpenuhi:
// - user sudah login melalui session, atau
// - header Authorization berisi Bearer token yang cocok dengan env ADMIN_API_KEY
func (server *Server) RequireAdminAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // cek token dari context (diset oleh AuthHeaderMiddleware) — prioritaskan token
        token := GetAuthTokenFromContext(r)
        if token != "" {
            if adminKey := os.Getenv("ADMIN_API_KEY"); adminKey != "" && token == adminKey {
                next.ServeHTTP(w, r)
                return
            }
        }

        // cek session + peran admin
        if IsLoggedIn(r) {
            user := server.CurrentUser(w, r)
            if user != nil && (user.Role == "admin" || user.Role == "superadmin") {
                next.ServeHTTP(w, r)
                return
            }
            // logged in but bukan admin
            http.Error(w, "forbidden", http.StatusForbidden)
            return
        }

        // unauthorized
        http.Error(w, "unauthorized", http.StatusUnauthorized)
    })
}

// GetAuthTokenFromContext mengembalikan token Authorization yang disimpan oleh middleware.
func GetAuthTokenFromContext(r *http.Request) string {
    v := r.Context().Value(ctxKeyAuthToken)
    if v == nil {
        return ""
    }
    if s, ok := v.(string); ok {
        return s
    }
    return ""
}

