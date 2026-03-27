package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// Base group for all API calls
	api := app.Group("/api")

	// ── 1. PUBLIC ROUTES ──────────────────────────────────────────────────────
	api.Post("/signup",         controllers.SignUp)
	api.Post("/login",          controllers.Login)
	api.Post("/forgot-password", controllers.ForgotPassword) 
	api.Post("/reset-password",  controllers.ResetPassword)  

	// ── 2. PROTECTED ROUTES ───────────────────────────────────────────────────
	auth := api.Group("/auth")
	auth.Use(middleware.RequireAuth)

	auth.Get("/user",   controllers.GetUser)
	auth.Post("/logout", controllers.Logout)
}