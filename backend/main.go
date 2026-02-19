package main

import (
	"backend/database"
	"backend/models"
	"backend/routes"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors" // Import the official middleware
)

func main() {
	app := fiber.New()

	log.Println("🚀 Starting Fiber server...")

	// Connect to database
	database.ConnectDB()

	// Auto-migrate user model
	if err := database.DB.
		Set("gorm:skipconstraint", true).
		AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("❌ Auto migration failed: %v", err)
	} else {
		log.Println("✅ Auto migration complete")
	}

	// ✅ CORRECTED CORS: Use the official middleware for security and reliability
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173, https://traders.kazini.africa", // Comma separated, NOT "||"
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// Setup routes
	routes.SetupRoutes(app)

	// Start server
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}