package controllers

import (
	"backend/database"
	"backend/models"

	"github.com/gofiber/fiber/v2"
)

// ── Admin ─────────────────────────────────────────────────────────────────────

func GetUsers(c *fiber.Ctx) error {
	var users []struct {
		ID        uint   `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Country   string `json:"country"`
		Role      string `json:"role"`
	}
	database.DB.Raw(`SELECT id, first_name, last_name, country, role FROM users ORDER BY created_at ASC`).Scan(&users)
	return c.JSON(users)
}

func GetUserCount(c *fiber.Ctx) error {
	var count int64
	database.DB.Raw(`SELECT COUNT(*) FROM users`).Scan(&count)
	return c.JSON(fiber.Map{"count": count})
}

// ── Student ───────────────────────────────────────────────────────────────────

func GetStudentDashboard(c *fiber.Ctx) error {
	user := c.Locals("user").(models.User)
	return c.JSON(fiber.Map{
		"message": "Welcome to your student dashboard",
		"user":    user.Safe(),
	})
}

func GetStudentStats(c *fiber.Ctx) error {
	user := c.Locals("user").(models.User)
	// TODO: replace with real queries when you have courses/progress tables
	return c.JSON(fiber.Map{
		"user_id":   user.ID,
		"courses":   0,
		"completed": 0,
		"progress":  0,
	})
}

// ── Trader ────────────────────────────────────────────────────────────────────

func GetTraderDashboard(c *fiber.Ctx) error {
	user := c.Locals("user").(models.User)
	return c.JSON(fiber.Map{
		"message": "Welcome to your trader dashboard",
		"user":    user.Safe(),
	})
}

func GetTraderStats(c *fiber.Ctx) error {
	user := c.Locals("user").(models.User)
	// TODO: replace with real queries when you have listings/transactions tables
	return c.JSON(fiber.Map{
		"user_id":  user.ID,
		"listings": 0,
		"revenue":  0,
		"trades":   0,
	})
}