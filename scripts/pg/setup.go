package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/maintenance"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=Europe/London"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&machines.Machine{})
	if err != nil {
		log.Fatalf(fmt.Sprintf("Database migration failed: %s", err.Error()))
	}

	err = db.AutoMigrate(&maintenance.Maintenance{})
	if err != nil {
		log.Fatalf(fmt.Sprintf("Database migration failed: %s", err.Error()))
	}
}
