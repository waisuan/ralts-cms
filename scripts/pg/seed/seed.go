package main

import (
	"context"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/testutils/factory"
	"time"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=Europe/London"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	r := machines.NewRepository(db)
	m := factory.BuildMachine()
	log.Println(fmt.Sprintf("Creating machine: %v", m.SerialNumber))
	_, err = r.Create(context.TODO(), m)
	if err != nil {
		log.Fatalf("Failed to create machine: %v", err)
	}
}
