package middleware

import (
	"backend/database"
	"backend/models"
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// getSecret fetches the JWT secret dynamically to ensure it isn't empty 
// if the middleware is initialized before the .env file is loaded.
func getSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Fallback for development; in production, this should log a critical error
		return []byte("default_secret_key") 
	}
	return []byte(secret)
}

// RequireAuth checks for Bearer token on protected routes
func RequireAuth(c *fiber.Ctx) error {
	// 1. Get Authorization Header
	auth := c.Get("Authorization")
	if auth == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing token",
		})
	}

	// 2. Validate Format (Bearer <token>)
	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token format",
		})
	}

	tokenStr := parts[1]

	// 3. Parse and Validate Token
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return getSecret(), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	// 4. Extract Claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token claims",
		})
	}

	// 5. Get User ID (sub) from claims
	// JWT numeric values are parsed as float64 by default
	subFloat, ok := claims["sub"].(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token subject",
		})
	}

	// 6. Fetch User from Database
	var user models.User
	if err := database.DB.First(&user, uint(subFloat)).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User no longer exists",
		})
	}

	// 7. Security: Strip password before passing user context
	user.Password = ""
	c.Locals("user", user)

	return c.Next()
}