package main

import (
	"log"
	"os" // Added to read environment variables

	"backend/database"
	"backend/models"
	"backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()

	log.Println("🚀 Starting Fiber server...")

	// 1. Connect to database
	// Make sure your DB_URL is set in Railway Variables!
	database.ConnectDB()

	// 2. Auto-migrate user model
	if err := database.DB.
		Set("gorm:skipconstraint", true).
		AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("❌ Auto migration failed: %v", err)
	} else {
		log.Println("✅ Auto migration complete")
	}

	// 3. CORS Configuration
app.Use(cors.New(cors.Config{
    AllowOrigins: "http://localhost:5173, https://traders.kazini.africa, https://afri-backend-production.up.railway.app",
    AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
    AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
    AllowCredentials: true,
}))

	// 4. Setup routes
	routes.SetupRoutes(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // Default for local development
	}

	log.Printf("📡 Server is listening on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}