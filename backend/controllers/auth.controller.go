package controllers

import (
	"backend/database"
	"backend/models"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Helper to get secret dynamically
func getJWTSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

// SignUp creates a new user and returns a JWT for auto-login
func SignUp(c *fiber.Ctx) error {
	var req models.User

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Basic Validation
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Please fill in all required fields"})
	}

	// Hash Password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to secure password"})
	}
	req.Password = string(hashed)

	// Create User with specific error handling
	// Avoid inserting empty phone_number (unique index conflict) by only including it when provided
	createData := map[string]interface{
		"first_name":     req.FirstName,
		"last_name":      req.LastName,
		"email":          req.Email,
		"country":        req.Country,
		"study_level":    req.StudyLevel,
		"field_of_study": req.FieldOfStudy,
		"year_of_study":  req.YearOfStudy,
		"learning_goals": req.LearningGoals,
		"password":       req.Password,
	}

	if pn := strings.TrimSpace(req.PhoneNumber); pn != "" {
		createData["phone_number"] = pn
	}

	if err := database.DB.Model(&models.User{}).Create(createData).Error; err != nil {
		// If the error contains "duplicate key", we know the email or phone exists
		if strings.Contains(err.Error(), "23505") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return c.Status(400).JSON(fiber.Map{"error": "This email or phone is already registered"})
		}
		return c.Status(400).JSON(fiber.Map{"error": "Registration failed: invalid data provided"})
	}

	// reload user to get ID and timestamps
	if err := database.DB.Where("email = ?", req.Email).First(&req).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch created user"})
	}

	// Auto-generate JWT
	claims := jwt.MapClaims{
		"sub": req.ID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(getJWTSecret())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error: token generation failed"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Welcome, " + req.FirstName + "!",
		"token":   signed,
		"user":    req.Safe(),
	})
}

// Login validates credentials and returns JWT
func Login(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request format"})
	}

	var user models.User
	if err := database.DB.Where("email = ?", strings.TrimSpace(body.Email)).First(&user).Error; err != nil {
		// We use a generic message for security so hackers don't know if the email exists
		return c.Status(401).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(getJWTSecret())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Login failed: could not issue token"})
	}

	return c.JSON(fiber.Map{
		"token": signed,
		"user":  user.Safe(),
	})
}

// Logout clears auth cookies
func Logout(c *fiber.Ctx) error {
	c.ClearCookie("afridauth")
	return c.JSON(fiber.Map{"message": "Successfully logged out"})
}