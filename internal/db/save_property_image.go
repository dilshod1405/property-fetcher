package db

import (
	"pf-service/internal/property"
	"gorm.io/gorm"
)

func SavePropertyImage(db *gorm.DB, img property.DjangoPropertyImage) error {
	return db.Create(&img).Error
}
