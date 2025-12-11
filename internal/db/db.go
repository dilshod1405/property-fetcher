package db

import (
	"log"
	"pf-service/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() *gorm.DB {
	dsn := config.AppConfig.PostgresDSN
	if dsn == "" {
		log.Fatal("POSTGRES_DSN is empty. Check .env")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("DB connect error:", err)
	}

	DB = db
	return db
}
