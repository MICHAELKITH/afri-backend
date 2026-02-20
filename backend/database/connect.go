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
    // Only load .env if it exists (avoids error logs in production)
    _ = godotenv.Load() 

    dsn := os.Getenv("DATABASE_URL")

    if dsn == "" {
        // Build manually ONLY if DATABASE_URL is missing
        host := os.Getenv("DB_HOST")
        user := os.Getenv("DB_USER")
        pass := os.Getenv("DB_PASSWORD")
        name := os.Getenv("DB_NAME")
        port := os.Getenv("DB_PORT")

        // If any of these are empty, GORM will produce that "user=" error
        dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", 
            host, user, pass, name, port)
    }

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        // This will now print the actual DSN (minus password) so you can see the error
        log.Fatalf("❌ DB Connection Error. Using DSN: %s | Error: %v", dsn, err)
    }

    DB = db
    log.Println("✅ Database connected successfully!")
}