package fakers

import (
	"time"

	"github.com/codeuiprogramming/e-commerce/app/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AdminFaker returns a single admin user for development/testing.
func AdminFaker(db *gorm.DB) *models.User {
    // generate bcrypt hash for the default admin password
    hashed, _ := bcrypt.GenerateFromPassword([]byte("adminpass"), bcrypt.DefaultCost)

    return &models.User{
        ID:            uuid.New().String(),
        FirstName:     "Admin",
        LastName:      "User",
        Email:         "admin@example.com",
        Password:      string(hashed),
        Role:          "admin",
        CreatedAt:     time.Now(),
    }
}
