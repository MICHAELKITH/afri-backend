package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
    _ = godotenv.Load() 

    dsn := os.Getenv("DATABASE_URL")

    if dsn == "" {
        log.Println("⚠️ DATABASE_URL not found, using individual DB_* variables")
        dsn = fmt.Sprintf(
            "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
            os.Getenv("DB_HOST"), os.Getenv("DB_USER"), 
            os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"),
        )
    } else {
        log.Println("🌐 Using DATABASE_URL for connection")
    }

    // DEBUG: This will show you exactly what host it's trying to hit
    // log.Printf("Attempting connection to: %s", dsn) 

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        // This is where your 'Connection Refused' error is currently coming from
        log.Fatalf("❌ Database connection failed: %v", err)
    }

    DB = db
    log.Println("✅ Database connected successfully!")
}