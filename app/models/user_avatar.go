package models

import (
	"time"

	"gorm.io/gorm"
)

type UserAvatar struct {
    ID        string         `gorm:"size:36;not null;uniqueIndex;primary_key"`
    UserID    string         `gorm:"size:36;index"`
    Path      string         `gorm:"size:255"`
    IsPrimary bool
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (a *UserAvatar) FindByUserID(db *gorm.DB, userID string) (*UserAvatar, error) {
    var avatar UserAvatar
    err := db.Debug().Where("user_id = ?", userID).First(&avatar).Error
    if err != nil {
        return nil, err
    }
    return &avatar, nil
}
