// +build integration

package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"pfservice/config"
	"pfservice/internal/area"
	"pfservice/internal/db"
	"pfservice/internal/httpclient"
	media "pfservice/internal/media_download"
	"pfservice/internal/property"
	"pfservice/internal/users"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestSync(t *testing.T) (*gorm.DB, *httptest.Server, string) {
	// Setup test database
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=test dbname=postgres_test port=5432 sslmode=disable"
	}

	testDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate tables
	err = testDB.AutoMigrate(
		&users.DjangoUser{},
		&property.DjangoProperty{},
		&property.DjangoPropertyTranslation{},
		&property.DjangoPropertyImage{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Clean up tables
	testDB.Exec("TRUNCATE TABLE core_app_customuser, core_app_property, core_app_property_translation, core_app_propertyimage RESTART IDENTITY CASCADE")

	// Create temporary media directory
	tmpMediaDir, err := os.MkdirTemp("", "pf-service-sync-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp media directory: %v", err)
	}

	os.Setenv("MEDIA_ROOT", tmpMediaDir)
	media.MediaRoot = tmpMediaDir

	// Create mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/token":
			handleMockTokenRequest(w, r)
		case "/users":
			handleMockUsersRequest(w, r)
		case "/listings":
			handleMockListingsRequest(w, r)
		default:
			// Handle image downloads
			if filepath.Ext(r.URL.Path) == ".jpg" || filepath.Ext(r.URL.Path) == ".png" {
				handleMockImageRequest(w, r)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}
	}))

	// Setup config
	config.AppConfig = &config.Config{
		PFAPIUrl:    mockServer.URL,
		PFAPIKey:    "test-api-key",
		PFAPISecret: "test-api-secret",
		PostgresDSN: dsn,
	}

	return testDB, mockServer, tmpMediaDir
}

func handleMockTokenRequest(w http.ResponseWriter, r *http.Request) {
	response := httpclient.TokenResponse{
		AccessToken: "mock-jwt-token",
		ExpiresIn:   3600,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleMockUsersRequest(w http.ResponseWriter, r *http.Request) {
	users := httpclient.PFUsersResponse{
		Data: []users.PFUser{
			{
				ID:     1001,
				Email:  "agent1@example.com",
				Mobile: "+971501234567",
				Status: "active",
				PublicProfile: &users.PFPublicProfile{
					ID:    1001,
					Name:  "Test Agent",
					Email: "agent1@example.com",
					Phone: "+971501234567",
				},
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func handleMockListingsRequest(w http.ResponseWriter, r *http.Request) {
	listings := httpclient.ListingsResponse{
		Results: []property.PFListing{
			{
				ID: "test-listing-001",
				Title: struct {
					En string `json:"en"`
				}{
					En: "Test Property",
				},
				Description: struct {
					En string `json:"en"`
				}{
					En: "Test Description",
				},
				Category:       "sale",
				FurnishingType: "furnished",
				Bedrooms: property.PFIntString{
					Value: 2,
				},
				Bathrooms: property.PFIntString{
					Value: 2,
				},
				Size: 1200.5,
				Location: struct {
					ID uint `json:"id"`
				}{
					ID: 3782,
				},
				AssignedTo: struct {
					ID int64 `json:"id"`
				}{
					ID: 1001,
				},
				Price: struct {
					Amounts struct {
						Sale int64 `json:"sale"`
					} `json:"amounts"`
				}{
					Amounts: struct {
						Sale int64 `json:"sale"`
					}{
						Sale: 1500000,
					},
				},
				Media: struct {
					Images []struct {
						Original struct {
							URL string `json:"url"`
						} `json:"original"`
					} `json:"images"`
				}{
					Images: []struct {
						Original struct {
							URL string `json:"url"`
						} `json:"original"`
					}{
						{
							Original: struct {
								URL string `json:"url"`
							}{
								URL: "/test-image-1.jpg",
							},
						},
						{
							Original: struct {
								URL string `json:"url"`
							}{
								URL: "/test-image-2.jpg",
							},
						},
					},
				},
				Reference: "REF-001",
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listings)
}

func handleMockImageRequest(w http.ResponseWriter, r *http.Request) {
	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(http.StatusOK)
	w.Write(jpegHeader)
}

func TestSyncDoesNotDeleteDatabaseRecords(t *testing.T) {
	testDB, mockServer, tmpMediaDir := setupTestSync(t)
	defer mockServer.Close()
	defer os.RemoveAll(tmpMediaDir)

	// Create existing property and image in database
	user := users.DjangoUser{
		Email:       "agent1@example.com",
		Phone:       "+971501234567",
		Role:        "agent",
		Password:    "!",
		IsActive:    true,
		IsStaff:     false,
		IsSuperuser: false,
	}
	savedUser, _ := db.SaveOrUpdateUser(testDB, user)
	userID := savedUser.ID

	prop := property.DjangoProperty{
		PfID:             "test-listing-001",
		UserID:           &userID,
		AreaID:           1,
		Bedrooms:         2,
		Bathrooms:        2,
		SquareSqft:       1200.5,
		Price:            1500000,
		StatusType:       "sale",
		ConstructionType: "furnished",
		Slug:             "test-listing-001",
		IsVisible:        true,
	}
	savedProp, _ := db.SaveOrUpdateProperty(testDB, prop, "Test Property", "Test Description")

	// Create existing image record in database
	existingImage := property.DjangoPropertyImage{
		PropertyID: savedProp.ID,
		Image:      "property_images/existing_image.jpg",
	}
	testDB.Create(&existingImage)
	existingImageID := existingImage.ID

	// Count records before sync
	var countBefore int64
	testDB.Model(&property.DjangoPropertyImage{}).Count(&countBefore)

	// Run sync (simulate main function logic)
	// Override config to use mock server
	oldConfig := config.AppConfig
	config.AppConfig = &config.Config{
		PFAPIUrl:    mockServer.URL,
		PFAPIKey:    "test-api-key",
		PFAPISecret: "test-api-secret",
		PostgresDSN: os.Getenv("TEST_POSTGRES_DSN"),
	}
	if config.AppConfig.PostgresDSN == "" {
		config.AppConfig.PostgresDSN = "host=localhost user=postgres password=test dbname=postgres_test port=5432 sslmode=disable"
	}
	defer func() {
		config.AppConfig = oldConfig
	}()

	token, _ := httpclient.GetJWTToken()
	allPFUsers, _ := httpclient.FetchAllUsers(token)
	listResp, _ := httpclient.FetchListings(token, 1)

	for _, listing := range listResp.Results {
		var pfAgent *users.PFUser
		for _, u := range allPFUsers {
			if u.PublicProfile != nil && u.PublicProfile.ID == listing.AssignedTo.ID {
				pfAgent = &u
				break
			}
		}

		if pfAgent == nil {
			continue
		}

		djUser := pfAgent.ToDjangoUser()
		savedUser, _ := db.SaveOrUpdateUser(testDB, djUser)
		userPointer := &savedUser.ID
		areaID := area.MapPFToDjangoArea(listing.Location.ID)
		prop := listing.ToDjangoProperty(userPointer, areaID)
		savedProp, _ := db.SaveOrUpdateProperty(testDB, prop, listing.Title.En, listing.Description.En)
		propIDuint := savedProp.ID

		// Download and save images
		for idx, img := range listing.Media.Images {
			url := img.Original.URL
			if url == "" {
				continue
			}

			fullURL := mockServer.URL + url
			localPath, err := media.DownloadImage(fullURL, propIDuint, idx)
			if err != nil {
				continue
			}

			if !media.ImageExists(localPath) {
				continue
			}

			// Check if already exists
			var existingImg property.DjangoPropertyImage
			err = testDB.Where("property_id = ? AND image = ?", propIDuint, localPath).First(&existingImg).Error
			if err == nil {
				continue
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}

			testDB.Create(&property.DjangoPropertyImage{
				PropertyID: propIDuint,
				Image:      localPath,
			})
		}
	}

	// Count records after sync
	var countAfter int64
	testDB.Model(&property.DjangoPropertyImage{}).Count(&countAfter)

	// Verify existing record still exists
	var stillExists property.DjangoPropertyImage
	err := testDB.Where("id = ?", existingImageID).First(&stillExists).Error
	if err != nil {
		t.Errorf("Existing image record was deleted! ID: %d", existingImageID)
	}

	// Verify count increased (new images added, but old ones not deleted)
	if countAfter < countBefore {
		t.Errorf("Database records decreased! Before: %d, After: %d", countBefore, countAfter)
	}
}

func TestSyncImageExistenceCheck(t *testing.T) {
	testDB, mockServer, tmpMediaDir := setupTestSync(t)
	defer mockServer.Close()
	defer os.RemoveAll(tmpMediaDir)

	// Create property
	user := users.DjangoUser{
		Email:       "agent1@example.com",
		Phone:       "+971501234567",
		Role:        "agent",
		Password:    "!",
		IsActive:    true,
		IsStaff:     false,
		IsSuperuser: false,
	}
	savedUser, _ := db.SaveOrUpdateUser(testDB, user)
	userID := savedUser.ID

	prop := property.DjangoProperty{
		PfID:             "test-listing-001",
		UserID:           &userID,
		AreaID:           1,
		Bedrooms:         2,
		Bathrooms:        2,
		SquareSqft:       1200.5,
		Price:            1500000,
		StatusType:       "sale",
		ConstructionType: "furnished",
		Slug:             "test-listing-001",
		IsVisible:        true,
	}
	savedProp, _ := db.SaveOrUpdateProperty(testDB, prop, "Test Property", "Test Description")

	// Create image record pointing to non-existent file
	missingImage := property.DjangoPropertyImage{
		PropertyID: savedProp.ID,
		Image:      "property_images/missing_file.jpg",
	}
	testDB.Create(&missingImage)
	missingImageID := missingImage.ID

	// Verify file doesn't exist
	if media.ImageExists(missingImage.Image) {
		t.Error("Test file should not exist")
	}

	// Check missing images
	missingImages, err := db.CheckMissingImages(testDB)
	if err != nil {
		t.Fatalf("Failed to check missing images: %v", err)
	}

	// Should find our missing image
	found := false
	for _, missing := range missingImages {
		if missing.ImageID == missingImageID {
			found = true
			break
		}
	}

	if !found {
		t.Error("Missing image check should find the non-existent file")
	}

	// Verify database record still exists (not deleted)
	var stillExists property.DjangoPropertyImage
	err = testDB.Where("id = ?", missingImageID).First(&stillExists).Error
	if err != nil {
		t.Errorf("Image record was deleted during check! ID: %d", missingImageID)
	}
}
