package main

import (
	"log"
	"os"

	"backend/database"
	"backend/models"
	"backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	log.Println("🚀 Initializing system dependencies...")

	// 1. Force database connection FIRST before doing anything else
	database.ConnectDB()

	// 2. Defensive Check: Ensure the GORM pointer isn't nil
	if database.DB == nil {
		log.Fatal("❌ Critical Error: Database connection object (DB) is nil! Shutting down.")
	}

	// 3. Auto-migrate user model safely
	log.Println("🔄 Running database migrations...")
	if err := database.DB.
		Set("gorm:skipconstraint", true).
		AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("❌ Auto migration failed: %v", err)
	}
	log.Println("✅ Database migration complete.")

	// 4. Initialize Fiber App
	app := fiber.New()

	// 5. CORS Configuration
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173, https://traders.kazini.africa, https://afri-backend-production.up.railway.app",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// 6. Setup routes (Your user count logic route should be registered safely inside here)
	log.Println("🛣️ Registering application routes...")
	routes.SetupRoutes(app)

	// 7. Dynamic Port Binding for Railway
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // Fallback for local machine development
	}

	log.Printf("📡 Afri-Backend live and listening on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("❌ Server failed to bind or stay open: %v", err)
	}
}