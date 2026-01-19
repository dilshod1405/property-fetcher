package main

import (
	"errors"
	"log"
	"pfservice/config"
	"pfservice/internal/area"
	"pfservice/internal/db"
	"pfservice/internal/httpclient"
	media "pfservice/internal/media_download"
	"pfservice/internal/property"
	"pfservice/internal/reporting"
	"pfservice/internal/users"

	"gorm.io/gorm"
)

func main() {
	config.LoadConfig()
	log.Println("PF SYNC STARTED...")

	// Initialize statistics
	stats := reporting.ReportStats{
		Date: reporting.GetTashkentTime(),
	}

	dbConn := db.Connect()

	// Check for missing images (read-only check, no deletion)
	log.Println("Checking existing images...")
	missingImages, err := db.CheckMissingImages(dbConn)
	if err != nil {
		log.Printf("Warning: Failed to check missing images: %v", err)
	} else if len(missingImages) > 0 {
		log.Printf("Found %d missing images in database. Will attempt to re-download during sync...", len(missingImages))
	} else {
		log.Println("All existing images verified - no missing files found")
	}

	token, err := httpclient.GetJWTToken()
	if err != nil {
		log.Fatal("Token error:", err)
	}

	allPFUsers, err := httpclient.FetchAllUsers(token)
	if err != nil {
		log.Fatal("PF Users fetch error:", err)
	}

	listResp, err := httpclient.FetchListings(token, 1)
	if err != nil {
		log.Fatal("PF Listings error:", err)
	}

	for _, listing := range listResp.Results {
		log.Println("Processing:", listing.ID)

		// FIND AGENT
		var pfAgent *users.PFUser
		for _, u := range allPFUsers {
			if u.PublicProfile != nil && u.PublicProfile.ID == listing.AssignedTo.ID {
				pfAgent = &u
				break
			}
		}

		if pfAgent == nil {
			log.Println("Agent not found:", listing.AssignedTo.ID)
			continue
		}

		// SAVE USER
		djUser := pfAgent.ToDjangoUser()

		// Check if user exists before saving to track creation/update
		var existingUser users.DjangoUser
		userExists := dbConn.Where("email = ?", djUser.Email).First(&existingUser).Error == nil

		savedUser, err := db.SaveOrUpdateUser(dbConn, djUser)
		if err != nil {
			log.Println("User save error:", err)
			stats.Errors++
			continue
		}

		// Track user creation/update
		if !userExists {
			stats.UsersCreated++
		} else {
			stats.UsersUpdated++
		}

		userPointer := &savedUser.ID

		// AREA
		areaID := area.MapPFToDjangoArea(listing.Location.ID)

		// CREATE/UPDATE PROPERTY
		prop := listing.ToDjangoProperty(userPointer, areaID)

		// Check if property exists before saving to track creation/update
		var existingProp property.DjangoProperty
		propExists := dbConn.Where("pf_id = ?", prop.PfID).First(&existingProp).Error == nil

		savedProp, changed := db.SaveOrUpdateProperty(
			dbConn,
			prop,
			listing.Title.En,
			listing.Description.En,
		)

		// Track property creation/update
		if !propExists {
			stats.PropertiesCreated++
		} else if changed {
			stats.PropertiesUpdated++
		}

		// PROPERTY ID FOR IMAGES (uint, correct)
		propIDuint := savedProp.ID

		// Check existing images for this property and re-download missing ones
		var existingImages []property.DjangoPropertyImage
		dbConn.Where("property_id = ?", propIDuint).Find(&existingImages)
		for _, existingImg := range existingImages {
			if !media.ImageExists(existingImg.Image) {
				log.Printf("Existing image missing for property %d: %s. Attempting to re-download...", propIDuint, existingImg.Image)
				// Try to find matching URL in current listing and re-download
				for imgIdx, listingImg := range listing.Media.Images {
					if listingImg.Original.URL != "" {
						localPath, err := media.DownloadImage(listingImg.Original.URL, propIDuint, imgIdx)
						if err == nil && media.ImageExists(localPath) {
							// Update existing record with new path instead of creating duplicate
							existingImg.Image = localPath
							err = dbConn.Save(&existingImg).Error
							if err != nil {
								log.Printf("Failed to update image record for property %d: %v", propIDuint, err)
								stats.Errors++
							} else {
								log.Printf("Re-downloaded and updated missing image for property %d: %s", propIDuint, localPath)
								stats.ImagesDownloaded++
							}
							break // Found and downloaded one image
						}
					}
				}
			}
		}

		// SAVE NEW IMAGES
		// Debug: Log media structure to understand API response
		if len(listing.Media.Images) == 0 {
			log.Printf("No images found in listing %s (pf_id: %s)", listing.ID, prop.PfID)
		} else {
			log.Printf("Found %d images in listing %s", len(listing.Media.Images), listing.ID)
		}

		for idx, img := range listing.Media.Images {
			url := img.Original.URL
			if url == "" {
				log.Printf("Empty URL for image %d in listing %s", idx, listing.ID)
				continue
			}

			log.Printf("Downloading image %d for property %d: %s", idx+1, propIDuint, url)

			// DownloadImage already has retry logic built-in
			// It will retry up to IMAGE_DOWNLOAD_MAX_RETRIES times (default: 3)
			// Pass image index to create unique filenames
			localPath, err := media.DownloadImage(url, propIDuint, idx)
			if err != nil {
				log.Printf("Image download failed after retries for property %d, URL: %s, error: %v", propIDuint, url, err)
				stats.Errors++
				continue
			}

			// Verify image file actually exists before saving to database
			if !media.ImageExists(localPath) {
				log.Printf("Downloaded image file does not exist at %s, skipping database save", localPath)
				continue
			}

			// Check if this image path already exists in database for this property
			var existingImg property.DjangoPropertyImage
			err = dbConn.Where("property_id = ? AND image = ?", propIDuint, localPath).First(&existingImg).Error
			if err == nil {
				// Image already exists in database, skip
				continue
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				// Some other error occurred, log and continue
				log.Printf("Error checking existing image for property %d: %v", propIDuint, err)
				continue
			}

			err = db.SavePropertyImage(dbConn, property.DjangoPropertyImage{
				PropertyID: propIDuint,
				Image:      localPath,
			})
			if err != nil {
				log.Printf("Failed to save property image to database for property %d, path: %s, error: %v", propIDuint, localPath, err)
				stats.Errors++
				continue
			}

			stats.ImagesDownloaded++
		}
	}

	// Write report
	stats.Date = reporting.GetTashkentTime()
	if err := reporting.WriteReport(stats); err != nil {
		log.Printf("Warning: Failed to write report: %v", err)
	}

	log.Println("IMPORT FINISHED SUCCESSFULLY")
	log.Printf("Summary: Created %d properties, Updated %d properties, Downloaded %d images, Created %d users, Updated %d users, Errors: %d",
		stats.PropertiesCreated, stats.PropertiesUpdated, stats.ImagesDownloaded, stats.UsersCreated, stats.UsersUpdated, stats.Errors)
}
