package db

import (
	"propertyfinder/internal/property"
	"gorm.io/gorm"
)

func SaveOrUpdateProperty(
	dbConn *gorm.DB,
	prop property.DjangoProperty,
	title string,
	description string,
) property.DjangoProperty {

	var existing property.DjangoProperty

	err := dbConn.Where("pf_id = ?", prop.PfID).First(&existing).Error

	if err == nil {
		// UPDATE
		prop.ID = existing.ID

		dbConn.Model(&existing).Updates(prop)

		// TRANSLATION UPDATE
		dbConn.Model(&property.DjangoPropertyTranslation{}).
			Where("master_id = ? AND language_code = 'en'", existing.ID).
			Updates(map[string]interface{}{
				"title":       title,
				"description": description,
			})

		return existing
	}

	// CREATE NEW PROPERTY
	dbConn.Create(&prop)

	// INSERT TRANSLATION
	dbConn.Create(&property.DjangoPropertyTranslation{
		MasterID:     prop.ID,
		LanguageCode: "en",
		Title:        title,
		Description:  description,
	})

	return prop
}
