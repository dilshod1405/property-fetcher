package db

import (
	"pfservice/internal/property"
	"gorm.io/gorm"
	"errors"
)

func SaveOrUpdateProperty(
	db *gorm.DB,
	prop property.DjangoProperty,
	title string,
	desc string,
) (property.DjangoProperty, bool) {

	var existing property.DjangoProperty

	err := db.Where("pf_id = ?", prop.PfID).First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		db.Create(&prop)
		property.SaveEnglishTranslation(db, prop.ID, title, "", desc)
		return prop, true
	}

	if err != nil {
		panic(err)
	}

	updates := map[string]interface{}{}

	if existing.Price != prop.Price {
		updates["price"] = prop.Price
	}
	if existing.Bedrooms != prop.Bedrooms {
		updates["bedrooms"] = prop.Bedrooms
	}
	if existing.Bathrooms != prop.Bathrooms {
		updates["bathrooms"] = prop.Bathrooms
	}
	if existing.SquareSqft != prop.SquareSqft {
		updates["square_sqft"] = prop.SquareSqft
	}
	if existing.StatusType != prop.StatusType {
		updates["status_type"] = prop.StatusType
	}
	if existing.ConstructionType != prop.ConstructionType {
		updates["construction_type"] = prop.ConstructionType
	}
	if !existing.IsVisible {
		updates["is_visible"] = true
	}

	changed := len(updates) > 0

	if changed {
		db.Model(&existing).Updates(updates)
	}

	property.SaveEnglishTranslation(db, existing.ID, title, "", desc)

	return existing, changed
}

