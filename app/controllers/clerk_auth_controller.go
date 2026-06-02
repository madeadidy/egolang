package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"

	"github.com/codeuiprogramming/e-commerce/app/models"
	"github.com/google/uuid"
)

// ClaimClerk accepts a minimal Clerk user payload from the frontend and
// creates or finds a corresponding local user, then sets the server session.
// NOTE: This is a development-friendly flow. In production you MUST verify the
// Clerk session token server-side using Clerk's SDK and the CLERK_SECRET_KEY.
func (server *Server) ClaimClerk(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		SessionID string `json:"sessionId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// debug log for incoming claims
	fmt.Println("ClaimClerk called with payload:", payload)

	// If CLERK_SECRET_KEY is set, perform server-side verification of the Clerk session id.
	clerkSecret := os.Getenv("CLERK_SECRET_KEY")
	if clerkSecret != "" {
		if payload.SessionID == "" {
			http.Error(w, "missing sessionId for server-side verification", http.StatusBadRequest)
			return
		}

		// verify session via Clerk REST API
		req, err := http.NewRequest("GET", fmt.Sprintf("https://api.clerk.dev/v1/sessions/%s", payload.SessionID), nil)
		if err != nil {
			fmt.Println("ClaimClerk: failed to build verification request:", err)
			http.Error(w, "failed to verify session", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Authorization", "Bearer "+clerkSecret)
		req.Header.Set("Accept", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("ClaimClerk: verification request failed:", err)
			http.Error(w, "failed to verify session", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			fmt.Println("ClaimClerk: Clerk verification failed, status:", resp.StatusCode, "body:", string(body))
			http.Error(w, "invalid clerk session", http.StatusUnauthorized)
			return
		}

		// Optional: we could parse the response body to verify user id/email match payload.
		fmt.Println("ClaimClerk: Clerk session verified for sessionId", payload.SessionID)
	}

	if payload.ID == "" || payload.Email == "" {
		http.Error(w, "missing id or email", http.StatusBadRequest)
		return
	}

	// try find existing user
	userModel := models.User{}
	existing, err := userModel.FindByID(server.DB, payload.ID)
	if err == nil && existing != nil {
		session, _ := store.Get(r, sessionUser)
		session.Values["id"] = existing.ID
		// ensure cookie options are explicit for this response
		session.Options = &sessions.Options{
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			MaxAge:   86400,
			SameSite: http.SameSiteLaxMode,
		}
		if err := session.Save(r, w); err != nil {
			fmt.Println("ClaimClerk: failed to save session for existing user:", err)
		} else {
			fmt.Println("ClaimClerk: session saved for existing user", existing.ID)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{ "code": 200, "message": "ok", "user": existing })
		return
	}

	// create a quick local user record using Clerk-provided id as primary key
	// generate a random password (hashed) using MakePassword helper
	randomPass := uuid.New().String() + time.Now().Format("20060102150405")
	hashed, _ := MakePassword(randomPass)

	newUser := &models.User{
		ID:        payload.ID,
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		Password:  hashed,
	}

	// ensure default role
	newUser.Role = "user"

	created, err := userModel.CreateUser(server.DB, newUser)
	if err != nil {
		// creation failed
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	session, _ := store.Get(r, sessionUser)
	session.Values["id"] = created.ID
	// ensure cookie options are explicit when creating session
	session.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
	}
	if err := session.Save(r, w); err != nil {
		fmt.Println("ClaimClerk: failed to save session for new user:", err)
	} else {
		fmt.Println("ClaimClerk: session saved for new user", created.ID)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{ "code": 200, "message": "created", "user": created })
}
