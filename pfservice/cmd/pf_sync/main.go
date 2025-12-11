package main

import (
	"log"
	"pfservice/config"
	"pfservice/internal/area"
	"pfservice/internal/db"
	"pfservice/internal/httpclient"
	"pfservice/internal/media_download"
	"pfservice/internal/property"
	"pfservice/internal/users"
)

func main() {
	config.LoadConfig()
	log.Println("PF SYNC STARTED...")

	dbConn := db.Connect()

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

		// Convert savedUser.ID (uint) to *uint
		userPointer := &savedUser.ID

		// AREA
		areaID := area.MapPFToDjangoArea(listing.Location.ID)

		// CREATE/UPDATE PROPERTY
		prop := listing.ToDjangoProperty(userPointer, areaID)

		savedProp := db.SaveOrUpdateProperty(dbConn, prop, listing.Title.En, listing.Description.En)

		// PROPERTY ID FOR IMAGES (uint, correct)
		propIDuint := savedProp.ID

		// SAVE IMAGES
		for _, img := range listing.Media.Images {
			url := img.Original.URL
			if url == "" {
				continue
			}

			localPath, err := media.DownloadImage(url, propIDuint)
			if err != nil {
				log.Println("Image download error:", err)
				continue
			}

			db.SavePropertyImage(dbConn, property.DjangoPropertyImage{
				PropertyID: propIDuint,
				Image:      localPath,
			})
		}
	}

	log.Println("IMPORT FINISHED SUCCESSFULLY")
}
