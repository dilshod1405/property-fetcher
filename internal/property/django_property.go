package property

import (
	"strings"
	"time"
)

type DjangoProperty struct {
	ID               uint      `gorm:"primaryKey;autoIncrement"`
	PfID             string    `gorm:"column:pf_id"`
	UserID           *uint     `gorm:"column:user_id"`
	AreaID           uint      `gorm:"column:area_id"`
	Longitude        float64   `gorm:"column:longitude"`
	Latitude         float64   `gorm:"column:latitude"`
	Bathrooms        int       `gorm:"column:bathrooms"`
	Bedrooms         int       `gorm:"column:bedrooms"`
	SquareSqft       float64   `gorm:"column:square_sqft"`
	Price            int64     `gorm:"column:price"`
	StatusType       string    `gorm:"column:status_type"`
	ConstructionType string    `gorm:"column:construction_type"`
	Slug             string    `gorm:"column:slug"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
	IsVisible        bool      `gorm:"column:is_visible"`
}

func (DjangoProperty) TableName() string {
	return "core_app_property"
}

type DjangoPropertyTranslation struct {
	ID           uint   `gorm:"primaryKey;autoIncrement"`
	MasterID     uint   `gorm:"column:master_id"`
	LanguageCode string `gorm:"column:language_code"`
	Title        string `gorm:"column:title"`
	Description  string `gorm:"column:description"`
}

func (DjangoPropertyTranslation) TableName() string {
	return "core_app_property_translation"
}

func (p PFListing) ToDjangoProperty(userID *uint, areaID uint) DjangoProperty {
	now := time.Now()

	status := p.Category
	if status == "" {
		status = "all"
	}

	construction := p.FurnishingType
	if construction == "" {
		construction = "apartment"
	}

	return DjangoProperty{
		PfID:             p.ID,
		UserID:           userID,
		AreaID:           areaID,
		Bedrooms:         p.Bedrooms.Value,
		Bathrooms:        p.Bathrooms.Value,
		SquareSqft:       p.Size,
		Price:            p.Price.Amounts.Sale,
		StatusType:       status,
		ConstructionType: construction,
		Slug:             strings.ToLower(p.ID),
		IsVisible:        true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}
