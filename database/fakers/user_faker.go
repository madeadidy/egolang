package fakers

import (
	"time"

	"github.com/bxcodec/faker/v4"
	"github.com/codeuiprogramming/e-commerce/app/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func UserFaker(db *gorm.DB) *models.User {
	return &models.User{
		ID:            uuid.New().String(),
		FirstName:     faker.FirstName(),
		LastName:      faker.LastName(),
		Email:         faker.Email(),
		Password:      "$2y$10$HLhnDLRuiUMIuTpE4gGOC.fCVaGlqOnQ3evm9DBw7CDt8cM33X/8W", //password
		RememberToken: "",
		CreatedAt:     time.Time{},
		UpdateAt: 	   time.Time{},
		DeleteAt:      gorm.DeletedAt{},
	}
}