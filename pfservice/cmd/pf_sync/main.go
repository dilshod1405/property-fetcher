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
	"pfservice/internal/users"

	"gorm.io/gorm"
)

func main() {
	config.LoadConfig()
	log.Println("PF SYNC STARTED...")

	dbConn := db.Connect()

	// Check for missing images and clean up database records
	log.Println("Checking existing images...")
	missingImages, err := db.CheckMissingImages(dbConn)
	if err != nil {
		log.Printf("Warning: Failed to check missing images: %v", err)
	} else if len(missingImages) > 0 {
		log.Printf("Found %d missing images in database. Cleaning up invalid records...", len(missingImages))
		for _, missing := range missingImages {
			err := db.DeletePropertyImage(dbConn, missing.ImageID)
			if err != nil {
				log.Printf("Failed to delete missing image record %d: %v", missing.ImageID, err)
			}
		}
		log.Printf("Cleaned up %d invalid image records", len(missingImages))
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
		savedUser, err := db.SaveOrUpdateUser(dbConn, djUser)
		if err != nil {
			log.Println("User save error:", err)
			continue
		}

		userPointer := &savedUser.ID

		// AREA
		areaID := area.MapPFToDjangoArea(listing.Location.ID)

		// CREATE/UPDATE PROPERTY
		prop := listing.ToDjangoProperty(userPointer, areaID)

		savedProp, _ := db.SaveOrUpdateProperty(
			dbConn,
			prop,
			listing.Title.En,
			listing.Description.En,
		)

		// PROPERTY ID FOR IMAGES (uint, correct)
		propIDuint := savedProp.ID

		// SAVE IMAGES
		for _, img := range listing.Media.Images {
			url := img.Original.URL
			if url == "" {
				continue
			}

			// DownloadImage already has retry logic built-in
			// It will retry up to IMAGE_DOWNLOAD_MAX_RETRIES times (default: 3)
			localPath, err := media.DownloadImage(url, propIDuint)
			if err != nil {
				log.Printf("Image download failed after retries for property %d, URL: %s, error: %v", propIDuint, url, err)
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
				continue
			}
		}
	}

	log.Println("IMPORT FINISHED SUCCESSFULLY")
}
