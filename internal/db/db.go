package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	
	"district-friends/internal/models"
)

func InitDB() *gorm.DB {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		host, user, password, dbname, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Friendship{},
		&models.ChatRoom{},
		&models.ChatRoomMember{},
		&models.Message{},
		&models.Event{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database connection established")

	// Seed mock data
	seedEvents(db)

	return db
}

func seedEvents(db *gorm.DB) {
	var count int64
	db.Model(&models.Event{}).Count(&count)
	if count == 0 {
		events := []models.Event{
			{Title: "Inception 10th Anniversary Screening", Description: "A mind-bending thriller.", Type: "Movie", Location: "Grand Cinema", Date: time.Now().Add(24 * time.Hour * 7), Price: 15.00},
			{Title: "Coldplay Live", Description: "Music of the Spheres World Tour.", Type: "Concert", Location: "Wembley Stadium", Date: time.Now().Add(24 * time.Hour * 30), Price: 150.00},
			{Title: "Gourmet Tasting Night", Description: "A 7-course meal by Chef Ramsay.", Type: "Dining", Location: "The French Laundry", Date: time.Now().Add(24 * time.Hour * 14), Price: 300.00},
			{Title: "Comedy Central Live", Description: "Standup comedy featuring top comedians.", Type: "Show", Location: "Laugh Factory", Date: time.Now().Add(24 * time.Hour * 3), Price: 45.00},
		}
		db.Create(&events)
		log.Println("Seeded events database with mock data.")
	}
}
