package db

import (
	"errors"
	"pf-service/internal/users"

	"gorm.io/gorm"
)

func SaveOrUpdateUser(db *gorm.DB, u users.DjangoUser) (users.DjangoUser, error) {
	var existing users.DjangoUser

	err := db.Where("email = ?", u.Email).First(&existing).Error
	if err == nil {
		// Update existing
		u.ID = existing.ID
		u.CreatedAt = existing.CreatedAt
		return u, db.Save(&u).Error
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return u, err
	}

	// Create new
	return u, db.Create(&u).Error
}


