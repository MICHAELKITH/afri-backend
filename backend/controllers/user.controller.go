package controllers

import (
	"backend/database"
	"backend/models"

	"github.com/gofiber/fiber/v2"
)

// ── Admin ─────────────────────────────────────────────────────────────────────

// GetUsers — now returns all fields the leaderboard needs.
// Keep this admin-only for the full list with emails.
func GetUsers(c *fiber.Ctx) error {
	var users []struct {
		ID        uint   `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Country   string `json:"country"`
		Role      string `json:"role"`
		CreatedAt string `json:"created_at"`
	}
	result := database.DB.Raw(`
		SELECT id, first_name, last_name, email, country, role, created_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at ASC
	`).Scan(&users)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch users"})
	}
	return c.JSON(users)
}

func GetUserCount(c *fiber.Ctx) error {
	var count int64
	database.DB.Raw(`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&count)
	return c.JSON(fiber.Map{"count": count})
}

// GetPublicStats — no auth required, safe aggregate counts for the dashboard.
func GetPublicStats(c *fiber.Ctx) error {
	var result struct {
		UserCount    int64 `json:"user_count"`
		TraderCount  int64 `json:"trader_count"`
		StudentCount int64 `json:"student_count"`
		CountryCount int64 `json:"country_count"`
	}
	database.DB.Raw(`SELECT COUNT(*) AS user_count FROM users WHERE deleted_at IS NULL`).Scan(&result)
	database.DB.Raw(`SELECT COUNT(*) AS trader_count FROM users WHERE role = 'trader' AND deleted_at IS NULL`).Scan(&result)
	database.DB.Raw(`SELECT COUNT(*) AS student_count FROM users WHERE role = 'student' AND deleted_at IS NULL`).Scan(&result)
	database.DB.Raw(`SELECT COUNT(DISTINCT country) AS country_count FROM users WHERE country != '' AND deleted_at IS NULL`).Scan(&result)
	return c.JSON(result)
}

// GetPublicLeaderboard — auth required (any role), but no email exposed.
func GetPublicLeaderboard(c *fiber.Ctx) error {
	var users []struct {
		ID        uint   `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Country   string `json:"country"`
		Role      string `json:"role"`
		CreatedAt string `json:"created_at"`
	}
	result := database.DB.Raw(`
		SELECT id, first_name, last_name, country, role, created_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at ASC
	`).Scan(&users)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch leaderboard"})
	}
	return c.JSON(users)
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
	return c.JSON(fiber.Map{
		"user_id":  user.ID,
		"listings": 0,
		"revenue":  0,
		"trades":   0,
	})
}