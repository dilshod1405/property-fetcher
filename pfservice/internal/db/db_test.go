package db

import (
	"os"
	"pfservice/internal/property"
	"pfservice/internal/users"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=test dbname=postgres_test port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate tables
	err = db.AutoMigrate(
		&users.DjangoUser{},
		&property.DjangoProperty{},
		&property.DjangoPropertyTranslation{},
		&property.DjangoPropertyImage{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Clean up tables before test
	db.Exec("TRUNCATE TABLE core_app_customuser, core_app_property, core_app_property_translation, core_app_propertyimage RESTART IDENTITY CASCADE")

	return db
}

func TestSaveOrUpdateUser(t *testing.T) {
	db := setupTestDB(t)

	user := users.DjangoUser{
		Email:       "test@example.com",
		Phone:       "+1234567890",
		Avatar:      "avatar.jpg",
		Role:        "agent",
		Password:    "!",
		IsActive:    true,
		IsStaff:     false,
		IsSuperuser: false,
	}

	// Test create
	saved, err := SaveOrUpdateUser(db, user)
	if err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	if saved.ID == 0 {
		t.Error("User ID should be set after save")
	}
	if saved.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, saved.Email)
	}

	// Test update
	user.Phone = "+9876543210"
	updated, err := SaveOrUpdateUser(db, user)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	if updated.ID != saved.ID {
		t.Error("User ID should remain the same on update")
	}
	if updated.Phone != user.Phone {
		t.Errorf("Expected phone %s, got %s", user.Phone, updated.Phone)
	}
}

func TestSaveOrUpdateProperty(t *testing.T) {
	db := setupTestDB(t)

	// Create a user first
	user := users.DjangoUser{
		Email:       "agent@example.com",
		Phone:       "+1234567890",
		Role:        "agent",
		Password:    "!",
		IsActive:    true,
		IsStaff:     false,
		IsSuperuser: false,
	}
	savedUser, _ := SaveOrUpdateUser(db, user)
	userID := savedUser.ID

	prop := property.DjangoProperty{
		PfID:             "pf-123",
		UserID:           &userID,
		AreaID:           1,
		Bedrooms:         2,
		Bathrooms:        2,
		SquareSqft:       1200.5,
		Price:            500000,
		StatusType:       "sale",
		ConstructionType: "apartment",
		Slug:             "pf-123",
		IsVisible:        true,
	}

	// Test create
	saved, created := SaveOrUpdateProperty(db, prop, "Test Property", "Test Description")
	if !created {
		t.Error("Property should be created on first save")
	}
	if saved.ID == 0 {
		t.Error("Property ID should be set after save")
	}
	if saved.PfID != prop.PfID {
		t.Errorf("Expected PfID %s, got %s", prop.PfID, saved.PfID)
	}

	// Test update
	prop.Price = 600000
	prop.Bedrooms = 3
	updated, changed := SaveOrUpdateProperty(db, prop, "Updated Property", "Updated Description")
	if !changed {
		t.Error("Property should be marked as changed when price/bedrooms differ")
	}
	if updated.ID != saved.ID {
		t.Error("Property ID should remain the same on update")
	}
	if updated.Price != prop.Price {
		t.Errorf("Expected price %d, got %d", prop.Price, updated.Price)
	}

	// Test no change
	noChangeProp := property.DjangoProperty{
		PfID:             "pf-123",
		UserID:           &userID,
		AreaID:           1,
		Bedrooms:         3,
		Bathrooms:        2,
		SquareSqft:       1200.5,
		Price:            600000,
		StatusType:       "sale",
		ConstructionType: "apartment",
		Slug:             "pf-123",
		IsVisible:        true,
	}
	_, changed = SaveOrUpdateProperty(db, noChangeProp, "Updated Property", "Updated Description")
	if changed {
		t.Error("Property should not be marked as changed when values are the same")
	}
}

func TestSavePropertyImage(t *testing.T) {
	db := setupTestDB(t)

	// Create a user and property first
	user := users.DjangoUser{
		Email:       "agent@example.com",
		Phone:       "+1234567890",
		Role:        "agent",
		Password:    "!",
		IsActive:    true,
		IsStaff:     false,
		IsSuperuser: false,
	}
	savedUser, _ := SaveOrUpdateUser(db, user)
	userID := savedUser.ID

	prop := property.DjangoProperty{
		PfID:             "pf-123",
		UserID:           &userID,
		AreaID:           1,
		Bedrooms:         2,
		Bathrooms:        2,
		SquareSqft:       1200.5,
		Price:            500000,
		StatusType:       "sale",
		ConstructionType: "apartment",
		Slug:             "pf-123",
		IsVisible:        true,
	}
	savedProp, _ := SaveOrUpdateProperty(db, prop, "Test Property", "Test Description")

	img := property.DjangoPropertyImage{
		PropertyID: savedProp.ID,
		Image:      "property_images/test.jpg",
	}

	err := SavePropertyImage(db, img)
	if err != nil {
		t.Fatalf("Failed to save property image: %v", err)
	}

	// Verify image was saved
	var savedImg property.DjangoPropertyImage
	err = db.Where("property_id = ?", savedProp.ID).First(&savedImg).Error
	if err != nil {
		t.Fatalf("Failed to retrieve saved image: %v", err)
	}
	if savedImg.Image != img.Image {
		t.Errorf("Expected image path %s, got %s", img.Image, savedImg.Image)
	}
}
