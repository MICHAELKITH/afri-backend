package routes

import (
    "backend/controllers"
    "backend/middleware"

    "github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
    // Base group for all API calls
    api := app.Group("/api")

    // 1. PUBLIC ROUTES
    // These are attached directly to the 'api' group
    api.Post("/signup", controllers.SignUp)
    api.Post("/login", controllers.Login)

    // 2. PROTECTED ROUTES 
    // We create a nested group specifically for routes requiring auth.
    // By giving it a prefix like "/v1" or even just a distinct group object,
    // we prevent it from interfering with the sibling routes above.
    auth := api.Group("/auth") 
    auth.Use(middleware.RequireAuth)

    // These will now be accessible at /api/auth/user and /api/auth/logout
    auth.Get("/user", controllers.GetUser)
    auth.Post("/logout", controllers.Logout)
}