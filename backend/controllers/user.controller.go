package controllers

import (
	"backend/database"

	"github.com/gofiber/fiber/v2"
)

func GetUsers(c *fiber.Ctx) error {
    var users []struct {
        ID        uint   `json:"id"`
        FirstName string `json:"first_name"`
        LastName  string `json:"last_name"`
        Country   string `json:"country"`
    }
    database.DB.Raw(`SELECT id, first_name, last_name, country FROM users ORDER BY created_at ASC`).Scan(&users)
    return c.JSON(users)
}

func GetUserCount(c *fiber.Ctx) error {
	var count int64
	database.DB.Raw(`SELECT COUNT(*) FROM users`).Scan(&count)
	return c.JSON(fiber.Map{"count": count})
}