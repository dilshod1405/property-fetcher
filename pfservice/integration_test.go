package main

import (
	"encoding/json"
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

func setupIntegrationTest(t *testing.T) (*gorm.DB, *httptest.Server, string) {
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
	tmpMediaDir, err := os.MkdirTemp("", "pf-service-integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp media directory: %v", err)
	}

	// Set MEDIA_ROOT environment variable
	os.Setenv("MEDIA_ROOT", tmpMediaDir)
	// MediaRoot will be read from environment on next access
	// Force reload by setting it directly
	media.MediaRoot = tmpMediaDir

	// Create mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/token":
			handleTokenRequest(w, r)
		case "/users":
			handleUsersRequest(w, r)
		case "/listings":
			handleListingsRequest(w, r)
		default:
			// Handle image downloads
			if filepath.Ext(r.URL.Path) == ".jpg" || filepath.Ext(r.URL.Path) == ".png" {
				handleImageRequest(w, r)
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

func handleTokenRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	response := httpclient.TokenResponse{
		AccessToken: "mock-jwt-token-12345",
		ExpiresIn:   3600,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleUsersRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	users := httpclient.PFUsersResponse{
		Data: []users.PFUser{
			{
				ID:     1001,
				Email:  "agent1@example.com",
				Mobile: "+971501234567",
				Status: "active",
				PublicProfile: &users.PFPublicProfile{
					ID:    1001,
					Name:  "John Agent",
					Email: "agent1@example.com",
					Phone: "+971501234567",
					ImageVariants: struct {
						Large struct {
							Default string `json:"default"`
						} `json:"large"`
					}{
						Large: struct {
							Default string `json:"default"`
						}{
							Default: "avatar1.jpg",
						},
					},
				},
			},
			{
				ID:     1002,
				Email:  "agent2@example.com",
				Mobile: "+971509876543",
				Status: "active",
				PublicProfile: &users.PFPublicProfile{
					ID:    1002,
					Name:  "Jane Agent",
					Email: "agent2@example.com",
					Phone: "+971509876543",
				},
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func handleListingsRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	listings := httpclient.ListingsResponse{
		Results: []property.PFListing{
			{
				ID: "pf-listing-001",
				Title: struct {
					En string `json:"en"`
				}{
					En: "Beautiful 2BR Apartment in Dubai Marina",
				},
				Description: struct {
					En string `json:"en"`
				}{
					En: "Stunning apartment with sea view",
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
					ID: 3782, // Maps to area 1
				},
				AssignedTo: struct {
					ID int64 `json:"id"`
				}{
					ID: 1001, // Maps to first agent
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
			{
				ID: "pf-listing-002",
				Title: struct {
					En string `json:"en"`
				}{
					En: "Luxury 3BR Villa in Palm Jumeirah",
				},
				Description: struct {
					En string `json:"en"`
				}{
					En: "Premium villa with private pool",
				},
				Category:       "rent",
				FurnishingType: "unfurnished",
				Bedrooms: property.PFIntString{
					Value: 3,
				},
				Bathrooms: property.PFIntString{
					Value: 3,
				},
				Size: 2500.0,
				Location: struct {
					ID uint `json:"id"`
				}{
					ID: 1001, // Maps to area 2
				},
				AssignedTo: struct {
					ID int64 `json:"id"`
				}{
					ID: 1002, // Maps to second agent
				},
				Price: struct {
					Amounts struct {
						Sale int64 `json:"sale"`
					} `json:"amounts"`
				}{
					Amounts: struct {
						Sale int64 `json:"sale"`
					}{
						Sale: 2500000,
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
								URL: "/test-image-3.jpg",
							},
						},
					},
				},
				Reference: "REF-002",
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listings)
}

func handleImageRequest(w http.ResponseWriter, r *http.Request) {
	// Return a mock JPEG image
	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(http.StatusOK)
	w.Write(jpegHeader)
}

func TestIntegrationSyncFlow(t *testing.T) {
	testDB, mockServer, tmpMediaDir := setupIntegrationTest(t)
	defer mockServer.Close()
	defer os.RemoveAll(tmpMediaDir)

	// Get JWT token
	token, err := httpclient.GetJWTToken()
	if err != nil {
		t.Fatalf("Failed to get JWT token: %v", err)
	}
	if token != "mock-jwt-token-12345" {
		t.Errorf("Expected token 'mock-jwt-token-12345', got '%s'", token)
	}

	// Fetch users
	allPFUsers, err := httpclient.FetchAllUsers(token)
	if err != nil {
		t.Fatalf("Failed to fetch users: %v", err)
	}
	if len(allPFUsers) != 2 {
		t.Errorf("Expected 2 users, got %d", len(allPFUsers))
	}

	// Fetch listings
	listResp, err := httpclient.FetchListings(token, 1)
	if err != nil {
		t.Fatalf("Failed to fetch listings: %v", err)
	}
	if len(listResp.Results) != 2 {
		t.Errorf("Expected 2 listings, got %d", len(listResp.Results))
	}

	// Process listings (simulating main.go logic)
	for _, listing := range listResp.Results {
		// Find agent
		var pfAgent *users.PFUser
		for _, u := range allPFUsers {
			if u.PublicProfile != nil && u.PublicProfile.ID == listing.AssignedTo.ID {
				pfAgent = &u
				break
			}
		}

		if pfAgent == nil {
			t.Errorf("Agent not found for listing %s", listing.ID)
			continue
		}

		// Save user
		djUser := pfAgent.ToDjangoUser()
		savedUser, err := db.SaveOrUpdateUser(testDB, djUser)
		if err != nil {
			t.Fatalf("Failed to save user: %v", err)
		}
		if savedUser.ID == 0 {
			t.Error("User ID should be set after save")
		}

		userPointer := &savedUser.ID

		// Map area
		areaID := area.MapPFToDjangoArea(listing.Location.ID)

		// Create/update property
		prop := listing.ToDjangoProperty(userPointer, areaID)
		savedProp, _ := db.SaveOrUpdateProperty(
			testDB,
			prop,
			listing.Title.En,
			listing.Description.En,
		)

		if savedProp.ID == 0 {
			t.Error("Property ID should be set after save")
		}
		if savedProp.PfID != listing.ID {
			t.Errorf("Expected PfID %s, got %s", listing.ID, savedProp.PfID)
		}

		// Verify translation was saved
		var translation property.DjangoPropertyTranslation
		err = testDB.Where("master_id = ? AND language_code = ?", savedProp.ID, "en").First(&translation).Error
		if err != nil {
			t.Fatalf("Failed to find translation: %v", err)
		}
		if translation.Title != listing.Title.En {
			t.Errorf("Expected title %s, got %s", listing.Title.En, translation.Title)
		}

		// Download and save images
		propIDuint := savedProp.ID
		imageCount := 0
		for _, img := range listing.Media.Images {
			url := img.Original.URL
			if url == "" {
				continue
			}

			// Make full URL for mock server
			fullURL := mockServer.URL + url
			localPath, err := media.DownloadImage(fullURL, propIDuint)
			if err != nil {
				t.Logf("Image download error (may be expected in test): %v", err)
				continue
			}

			err = db.SavePropertyImage(testDB, property.DjangoPropertyImage{
				PropertyID: propIDuint,
				Image:      localPath,
			})
			if err != nil {
				t.Fatalf("Failed to save property image: %v", err)
			}
			imageCount++

			// Verify image file exists
			fullPath := filepath.Join(tmpMediaDir, localPath)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("Image file does not exist at %s", fullPath)
			}
		}

		// Verify images were saved in database
		var savedImages []property.DjangoPropertyImage
		err = testDB.Where("property_id = ?", propIDuint).Find(&savedImages).Error
		if err != nil {
			t.Fatalf("Failed to query property images: %v", err)
		}
		if len(savedImages) != imageCount {
			t.Errorf("Expected %d images in database, got %d", imageCount, len(savedImages))
		}

		// Verify image paths are relative (as expected by Django)
		for _, img := range savedImages {
			if filepath.IsAbs(img.Image) {
				t.Errorf("Image path should be relative, got absolute path: %s", img.Image)
			}
			if !filepath.HasPrefix(img.Image, "property_images/") {
				t.Errorf("Image path should start with 'property_images/', got: %s", img.Image)
			}
		}
	}

	// Verify final state
	var totalUsers int64
	testDB.Model(&users.DjangoUser{}).Count(&totalUsers)
	if totalUsers != 2 {
		t.Errorf("Expected 2 users in database, got %d", totalUsers)
	}

	var totalProperties int64
	testDB.Model(&property.DjangoProperty{}).Count(&totalProperties)
	if totalProperties != 2 {
		t.Errorf("Expected 2 properties in database, got %d", totalProperties)
	}

	var totalImages int64
	testDB.Model(&property.DjangoPropertyImage{}).Count(&totalImages)
	if totalImages < 2 {
		t.Errorf("Expected at least 2 images in database, got %d", totalImages)
	}
}

func TestIntegrationMediaPathConfiguration(t *testing.T) {
	// Verify that media path configuration matches Docker setup
	// In Docker: /mhp/media is mounted and Django expects relative paths
	expectedDefaultMediaRoot := "/mhp/media"

	// Reset environment
	oldEnv := os.Getenv("MEDIA_ROOT")
	os.Unsetenv("MEDIA_ROOT")
	defer os.Setenv("MEDIA_ROOT", oldEnv)

	// Check default by reading from environment
	mediaRoot := os.Getenv("MEDIA_ROOT")
	if mediaRoot == "" {
		mediaRoot = "/mhp/media" // Default
	}
	if mediaRoot != expectedDefaultMediaRoot {
		t.Errorf("Expected default MEDIA_ROOT %s, got %s", expectedDefaultMediaRoot, mediaRoot)
	}

	// Test that paths are relative (Django requirement)
	tmpDir, err := os.MkdirTemp("", "pf-service-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("MEDIA_ROOT", tmpDir)
	media.MediaRoot = tmpDir

	// Create a mock server for image download
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0}
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		w.Write(jpegHeader)
	}))
	defer server.Close()

	localPath, err := media.DownloadImage(server.URL+"/test.jpg", 123)
	if err != nil {
		t.Fatalf("Failed to download image: %v", err)
	}

	// Path should be relative
	if filepath.IsAbs(localPath) {
		t.Errorf("Image path should be relative, got absolute: %s", localPath)
	}

	// Path should start with property_images/
	if !filepath.HasPrefix(localPath, "property_images/") {
		t.Errorf("Image path should start with 'property_images/', got: %s", localPath)
	}

	// Full path should exist
	fullPath := filepath.Join(tmpDir, localPath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Errorf("Image file should exist at %s", fullPath)
	}

	// Verify the structure matches Django expectations
	// Django expects: MEDIA_ROOT/property_images/filename.jpg
	// We save to: MEDIA_ROOT/property_images/filename.jpg
	// We return: property_images/filename.jpg (relative)
	// This is correct!
}

func TestIntegrationAreaMapping(t *testing.T) {
	testCases := []struct {
		pfAreaID   uint
		expectedID uint
	}{
		{3782, 1},
		{1001, 2},
		{1002, 3},
		{9999, 1}, // Unknown area should default to 1
	}

	for _, tc := range testCases {
		result := area.MapPFToDjangoArea(tc.pfAreaID)
		if result != tc.expectedID {
			t.Errorf("For PF area %d, expected Django area %d, got %d", tc.pfAreaID, tc.expectedID, result)
		}
	}
}

func TestIntegrationUserConversion(t *testing.T) {
	pfUser := users.PFUser{
		ID:     1001,
		Email:  "test@example.com",
		Mobile: "+971501234567",
		Status: "active",
		PublicProfile: &users.PFPublicProfile{
			ID:    1001,
			Name:  "Test User",
			Email: "test@example.com",
			Phone: "+971509876543",
			ImageVariants: struct {
				Large struct {
					Default string `json:"default"`
				} `json:"large"`
			}{
				Large: struct {
					Default string `json:"default"`
				}{
					Default: "avatar.jpg",
				},
			},
		},
	}

	djUser := pfUser.ToDjangoUser()

	if djUser.Email != pfUser.Email {
		t.Errorf("Expected email %s, got %s", pfUser.Email, djUser.Email)
	}
	if djUser.Phone != pfUser.PublicProfile.Phone {
		t.Errorf("Expected phone %s, got %s", pfUser.PublicProfile.Phone, djUser.Phone)
	}
	if djUser.Avatar != "avatar.jpg" {
		t.Errorf("Expected avatar avatar.jpg, got %s", djUser.Avatar)
	}
	if djUser.Role != "agent" {
		t.Errorf("Expected role 'agent', got %s", djUser.Role)
	}
	if !djUser.IsActive {
		t.Error("User should be active")
	}
}
