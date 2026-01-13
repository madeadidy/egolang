package controllers

import (
	"encoding/json"
	"net/http"
	"time"

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
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
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
		session.Save(r, w)

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

	created, err := userModel.CreateUser(server.DB, newUser)
	if err != nil {
		// creation failed
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	session, _ := store.Get(r, sessionUser)
	session.Values["id"] = created.ID
	session.Save(r, w)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{ "code": 200, "message": "created", "user": created })
}
