package db

import (
	"errors"
	"pfservice/internal/users"
	"strings"
	"gorm.io/gorm"
)


func normalizeAvatar(avatar string) string {
	if avatar == "" {
		return ""
	}

	if strings.HasPrefix(avatar, "http://") || strings.HasPrefix(avatar, "https://") {
		return ""
	}

	return avatar
}


func SaveOrUpdateUser(db *gorm.DB, u users.DjangoUser) (users.DjangoUser, error) {
	var existing users.DjangoUser

	err := db.Where("email = ?", u.Email).First(&existing).Error
	if err == nil {
		u.ID = existing.ID
		u.CreatedAt = existing.CreatedAt

		u.Avatar = normalizeAvatar(u.Avatar)

		return u, db.Save(&u).Error
	}


	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return u, err
	}

	u.Avatar = normalizeAvatar(u.Avatar)
	return u, db.Create(&u).Error

}


