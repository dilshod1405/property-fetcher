package main

import (
	"log"
	"pfservice/config"
	"pfservice/internal/db"
	"pfservice/internal/httpclient"
	media "pfservice/internal/media_download"
	"pfservice/internal/property"
)

func main() {
	config.LoadConfig()
	log.Println("PF IMAGE REPAIR STARTED...")

	dbConn := db.Connect()

	// Check for missing images
	log.Println("Checking for missing images...")
	missingImages, err := db.CheckMissingImages(dbConn)
	if err != nil {
		log.Fatalf("Failed to check missing images: %v", err)
	}

	if len(missingImages) == 0 {
		log.Println("No missing images found. All images are present.")
		return
	}

	log.Printf("Found %d missing images. Starting repair...", len(missingImages))

	// Get JWT token
	token, err := httpclient.GetJWTToken()
	if err != nil {
		log.Fatalf("Token error: %v", err)
	}

	// Group missing images by property
	missingByProperty := make(map[uint][]db.MissingImageInfo)
	pfIDs := make(map[string]bool)
	for _, missing := range missingImages {
		missingByProperty[missing.PropertyID] = append(missingByProperty[missing.PropertyID], missing)
		pfIDs[missing.PfID] = true
	}

	// Convert pfIDs map to slice
	pfIDList := make([]string, 0, len(pfIDs))
	for pfID := range pfIDs {
		pfIDList = append(pfIDList, pfID)
	}

	log.Printf("Fetching listings for %d properties...", len(pfIDList))

	// Fetch all listings to get image URLs
	// We'll need to fetch all pages to find our properties
	allListings := make(map[string]property.PFListing)
	page := 1
	maxPages := 100 // Safety limit

	for page <= maxPages {
		listResp, err := httpclient.FetchListings(token, page)
		if err != nil {
			log.Printf("Error fetching listings page %d: %v", page, err)
			break
		}

		if len(listResp.Results) == 0 {
			break
		}

		for _, listing := range listResp.Results {
			allListings[listing.ID] = listing
		}

		// If we got less than 50 results, we're probably on the last page
		if len(listResp.Results) < 50 {
			break
		}

		page++
	}

	log.Printf("Fetched %d listings from API", len(allListings))

	// Process each property with missing images
	repairedCount := 0
	failedCount := 0

	for propertyID, missingList := range missingByProperty {
		// Get property to find pf_id
		var prop property.DjangoProperty
		err := dbConn.Where("id = ?", propertyID).First(&prop).Error
		if err != nil {
			log.Printf("Property %d not found in database, skipping", propertyID)
			failedCount += len(missingList)
			continue
		}

		// Find listing in fetched listings
		listing, found := allListings[prop.PfID]
		if !found {
			log.Printf("Listing %s (pf_id) not found in API response for property %d, skipping", prop.PfID, propertyID)
			failedCount += len(missingList)
			continue
		}

		// Get all image URLs from listing
		imageURLs := make([]string, 0)
		for _, img := range listing.Media.Images {
			if img.Original.URL != "" {
				imageURLs = append(imageURLs, img.Original.URL)
			}
		}

		if len(imageURLs) == 0 {
			log.Printf("No image URLs found for property %d (pf_id: %s), skipping", propertyID, prop.PfID)
			failedCount += len(missingList)
			continue
		}

		// Try to re-download images
		// We'll try to match by position or download all and let the system handle duplicates
		log.Printf("Repairing %d missing images for property %d (pf_id: %s)", len(missingList), propertyID, prop.PfID)

		for i, missing := range missingList {
			// Delete the missing image record from database
			err := db.DeletePropertyImage(dbConn, missing.ImageID)
			if err != nil {
				log.Printf("Failed to delete missing image record %d: %v", missing.ImageID, err)
			}

			// Try to download an image (use corresponding URL if available, otherwise use first available)
			urlIndex := i
			if urlIndex >= len(imageURLs) {
				urlIndex = 0 // Use first image if we don't have enough URLs
			}

			url := imageURLs[urlIndex]
			if url == "" {
				log.Printf("Empty URL for property %d, image %d, skipping", propertyID, i)
				failedCount++
				continue
			}

			// Download image
			localPath, err := media.DownloadImage(url, propertyID, i)
			if err != nil {
				log.Printf("Failed to download image for property %d, URL: %s, error: %v", propertyID, url, err)
				failedCount++
				continue
			}

			// Save to database
			err = db.SavePropertyImage(dbConn, property.DjangoPropertyImage{
				PropertyID: propertyID,
				Image:      localPath,
			})
			if err != nil {
				log.Printf("Failed to save image to database for property %d, path: %s, error: %v", propertyID, localPath, err)
				failedCount++
				continue
			}

			log.Printf("Successfully repaired image for property %d: %s", propertyID, localPath)
			repairedCount++
		}
	}

	log.Printf("REPAIR FINISHED: %d images repaired, %d failed", repairedCount, failedCount)
}

