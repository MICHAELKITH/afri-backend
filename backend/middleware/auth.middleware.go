package middleware

import (
	"backend/database"
	"backend/models"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func getSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("❌ JWT_SECRET environment variable is not set")
	}
	return []byte(secret)
}

func RequireAuth(c *fiber.Ctx) error {
	auth := c.Get("Authorization")
	if auth == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing token"})
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token format"})
	}

	token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return getSecret(), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token claims"})
	}

	subFloat, ok := claims["sub"].(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token subject"})
	}

	var user models.User
	if err := database.DB.First(&user, uint(subFloat)).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User no longer exists"})
	}

	user.Password = ""
	c.Locals("user", user)
	return c.Next()
}

// RequireRole restricts a route to users with a specific role.
// Must be used after RequireAuth (depends on "user" being set in Locals).
func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}
		if user.Role != role {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied — your account type cannot access this resource",
			})
		}
		return c.Next()
	}
}