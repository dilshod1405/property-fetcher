package db

import (
	"os"
	"path/filepath"
	media "pfservice/internal/media_download"
	"pfservice/internal/property"

	"gorm.io/gorm"
)

// MissingImageInfo contains information about a missing image
type MissingImageInfo struct {
	ImageID    uint
	PropertyID uint
	PfID       string
	ImagePath  string
}

// CheckMissingImages scans all property images in the database and checks if files exist
// Returns a list of missing images
func CheckMissingImages(db *gorm.DB) ([]MissingImageInfo, error) {
	var allImages []property.DjangoPropertyImage
	err := db.Find(&allImages).Error
	if err != nil {
		return nil, err
	}

	var missingImages []MissingImageInfo

	for _, img := range allImages {
		// Get full path to image file
		fullPath := filepath.Join(media.MediaRoot, img.Image)

		// Check if file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			// File doesn't exist, get property info to find pf_id
			var prop property.DjangoProperty
			err := db.Where("id = ?", img.PropertyID).First(&prop).Error
			if err != nil {
				// Property not found, skip
				continue
			}

			missingImages = append(missingImages, MissingImageInfo{
				ImageID:    img.ID,
				PropertyID: img.PropertyID,
				PfID:       prop.PfID,
				ImagePath:  img.Image,
			})
		}
	}

	return missingImages, nil
}

// GetAllPropertyImages returns all property images grouped by property ID
func GetAllPropertyImages(db *gorm.DB) (map[uint][]property.DjangoPropertyImage, error) {
	var allImages []property.DjangoPropertyImage
	err := db.Find(&allImages).Error
	if err != nil {
		return nil, err
	}

	imagesByProperty := make(map[uint][]property.DjangoPropertyImage)
	for _, img := range allImages {
		imagesByProperty[img.PropertyID] = append(imagesByProperty[img.PropertyID], img)
	}

	return imagesByProperty, nil
}

// GetPropertiesByPfIDs fetches properties by their PF IDs
func GetPropertiesByPfIDs(db *gorm.DB, pfIDs []string) ([]property.DjangoProperty, error) {
	var properties []property.DjangoProperty
	err := db.Where("pf_id IN ?", pfIDs).Find(&properties).Error
	return properties, err
}

// DeletePropertyImage deletes an image record from the database
func DeletePropertyImage(db *gorm.DB, imageID uint) error {
	return db.Delete(&property.DjangoPropertyImage{}, imageID).Error
}
