package db

import (
	"pfservice/internal/property"
	"gorm.io/gorm"
)

func SavePropertyImage(db *gorm.DB, img property.DjangoPropertyImage) error {
	return db.Create(&img).Error
}
