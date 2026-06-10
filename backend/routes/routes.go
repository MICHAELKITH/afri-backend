package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// ── Public ────────────────────────────────────────────────────────────────
	api.Post("/signup",          controllers.SignUp)
	api.Post("/login/student",   controllers.LoginStudent)
	api.Post("/login/trader",    controllers.LoginTrader)
	api.Post("/forgot-password", controllers.ForgotPassword)
	api.Post("/reset-password",  controllers.ResetPassword)
	api.Post("/check-email",     controllers.CheckEmail)

	// ── Authenticated (any role) ──────────────────────────────────────────────
	auth := api.Group("/auth", middleware.RequireAuth)
	auth.Post("/logout", controllers.Logout)

	// ── Student only ──────────────────────────────────────────────────────────
	student := api.Group("/student", middleware.RequireAuth, middleware.RequireRole("student"))
	student.Get("/dashboard", controllers.GetStudentDashboard)
	student.Get("/stats",     controllers.GetStudentStats)

	// ── Trader only ───────────────────────────────────────────────────────────
	trader := api.Group("/trader", middleware.RequireAuth, middleware.RequireRole("trader"))
	trader.Get("/dashboard", controllers.GetTraderDashboard)
	trader.Get("/stats",     controllers.GetTraderStats)

	// ── Admin only ────────────────────────────────────────────────────────────
	admin := api.Group("/admin", middleware.RequireAuth, middleware.RequireRole("admin"))
	admin.Get("/users",       controllers.GetUsers)
	admin.Get("/users/count", controllers.GetUserCount)
}