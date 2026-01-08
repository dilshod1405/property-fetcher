package main

import (
	"log"
	"os"
	"pfservice/config"
	"pfservice/internal/db"
)

func main() {
	config.LoadConfig()

	// Check if --repair flag is provided
	repairMode := false
	if len(os.Args) > 1 && os.Args[1] == "--repair" {
		repairMode = true
	}

	if repairMode {
		log.Println("PF IMAGE CHECK & REPAIR MODE")
		log.Println("This will check for missing images and attempt to repair them.")
		log.Println("Run 'pf_repair' command for full repair functionality.")
	} else {
		log.Println("PF IMAGE CHECK MODE")
		log.Println("Checking for missing images...")
	}

	dbConn := db.Connect()

	// Check for missing images
	missingImages, err := db.CheckMissingImages(dbConn)
	if err != nil {
		log.Fatalf("Failed to check missing images: %v", err)
	}

	if len(missingImages) == 0 {
		log.Println("✓ No missing images found. All images are present.")
		return
	}

	log.Printf("✗ Found %d missing images:", len(missingImages))

	// Group by property for better reporting
	missingByProperty := make(map[uint]int)
	pfIDMap := make(map[uint]string)
	for _, missing := range missingImages {
		missingByProperty[missing.PropertyID]++
		if pfIDMap[missing.PropertyID] == "" {
			pfIDMap[missing.PropertyID] = missing.PfID
		}
	}

	log.Printf("\nMissing images by property:")
	for propertyID, count := range missingByProperty {
		pfID := pfIDMap[propertyID]
		log.Printf("  Property ID %d (pf_id: %s): %d missing images", propertyID, pfID, count)
	}

	log.Printf("\nTo repair missing images, run: pf_repair")
}

