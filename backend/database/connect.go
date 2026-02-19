package database

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// 1. Load .env for LOCAL development only
	// We ignore the error because Railway provides variables via system env
	_ = godotenv.Load()

	// 2. Priority 1: Use the full DATABASE_URL (Railway's default)
	dsn := os.Getenv("DATABASE_URL")

	// 3. Priority 2: Fallback to individual variables (Local dev default)
	if dsn == "" {
		log.Println("⚠️ DATABASE_URL not found, constructing DSN from individual variables...")
		dsn = "host=" + os.Getenv("DB_HOST") +
			" user=" + os.Getenv("DB_USER") +
			" password=" + os.Getenv("DB_PASSWORD") +
			" dbname=" + os.Getenv("DB_NAME") +
			" port=" + os.Getenv("DB_PORT") +
			" sslmode=disable"
	}

	// 4. Connect to GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	DB = db
	log.Println("✅ Database connected successfully")
}