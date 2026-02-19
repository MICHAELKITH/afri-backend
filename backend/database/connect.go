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
    // 1. Load .env only for local development
    _ = godotenv.Load() 

    // 2. Try the single Railway-provided URL first
    dsn := os.Getenv("DATABASE_URL")

    // 3. Fallback to individual variables if DATABASE_URL is not found (local dev)
    if dsn == "" {
        dsn = fmt.Sprintf(
            "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
            os.Getenv("DB_HOST"), os.Getenv("DB_USER"), 
            os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"),
        )
    }

    // 4. Connect via GORM
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("❌ Database connection failed: %v", err)
    }

    DB = db
    log.Println("✅ Database connected via GORM")
}