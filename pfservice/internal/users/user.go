package users
import (
	"time"
)

type PFUser struct {
	ID        int64            `json:"id"`
	Email     string           `json:"email"`
	Mobile    string           `json:"mobile"`
	Status    string           `json:"status"`
	PublicProfile *PFPublicProfile `json:"publicProfile"`
}

type PFPublicProfile struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`


	ImageVariants struct {
		Large struct {
			Default string `json:"default"`
		} `json:"large"`
	} `json:"imageVariants"`
}

type DjangoUser struct {
	ID         uint   `gorm:"primaryKey"`
	Email      string
	Phone      string
	Avatar     string
	Role       string
	Password   string
	IsActive   bool   `gorm:"column:is_active"`
	IsStaff    bool   `gorm:"column:is_staff"`
	IsSuperuser bool  `gorm:"column:is_superuser"`

	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}



func (DjangoUser) TableName() string {
    return "core_app_customuser"
}


func (p PFUser) ToDjangoUser() DjangoUser {
	avatar := ""
	phone := p.Mobile
	now := time.Now()

	if p.PublicProfile != nil {
		if p.PublicProfile.ImageVariants.Large.Default != "" {
			avatar = p.PublicProfile.ImageVariants.Large.Default
		}

		if p.PublicProfile.Phone != "" {
			phone = p.PublicProfile.Phone
		}
	}

	return DjangoUser{
		Email:       p.Email,
		Phone:       phone,
		Avatar:      avatar,
		Role:        "agent",
		Password:    "!",        // 🔥 REQUIRED
		IsActive:    p.Status == "active",
		IsStaff:     false,      // 🔥 REQUIRED
		IsSuperuser: false,      // 🔥 REQUIRED
		CreatedAt:   now,        // 🔥 REQUIRED
		UpdatedAt:   now,        // 🔥 REQUIRED
	}
}
