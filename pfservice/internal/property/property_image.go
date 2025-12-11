package property

type DjangoPropertyImage struct {
	ID         uint   `gorm:"primaryKey"`
	PropertyID uint   `gorm:"column:property_id"`
	Image      string `gorm:"column:image"`
}

func (DjangoPropertyImage) TableName() string {
	return "core_app_propertyimage"
}
