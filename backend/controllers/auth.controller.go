package controllers

import (
	"backend/database"
	"backend/models"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ─── Helper ───────────────────────────────────────────────────────────────────
func getJWTSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

// ─── SignUp ───────────────────────────────────────────────────────────────────
func SignUp(c *fiber.Ctx) error {
	var req models.User

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Please fill in all required fields"})
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to secure password"})
	}
	req.Password = string(hashed)

	createData := map[string]interface{}{
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
		if strings.Contains(err.Error(), "23505") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return c.Status(400).JSON(fiber.Map{"error": "This email or phone is already registered"})
		}
		return c.Status(400).JSON(fiber.Map{"error": "Registration failed: invalid data provided"})
	}

	if err := database.DB.Where("email = ?", req.Email).First(&req).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch created user"})
	}

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

// ─── Login ────────────────────────────────────────────────────────────────────
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

// ─── Logout ───────────────────────────────────────────────────────────────────
func Logout(c *fiber.Ctx) error {
	c.ClearCookie("afridauth")
	return c.JSON(fiber.Map{"message": "Successfully logged out"})
}

// ─── ForgotPassword ───────────────────────────────────────────────────────────
// POST /api/forgot-password
// Always returns 200 whether or not the email exists (prevents enumeration).
func ForgotPassword(c *fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	if body.Email == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email is required"})
	}

	// Look up user — if not found, still return 200 (no enumeration)
	var user models.User
	if err := database.DB.Where("email = ?", body.Email).First(&user).Error; err != nil {
		return c.Status(200).JSON(fiber.Map{
			"message": "If that email is registered, you will receive a reset link shortly.",
		})
	}

	// Generate secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate reset token"})
	}
	resetToken := hex.EncodeToString(tokenBytes)
	expiresAt  := time.Now().Add(1 * time.Hour)

	// Save token + expiry to the user record
	if err := database.DB.Model(&user).Updates(map[string]interface{}{
		"password_reset_token":   resetToken,
		"password_reset_expires": expiresAt,
	}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save reset token"})
	}

	// Build reset link using FRONTEND_URL from .env
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, resetToken)

	// Send email asynchronously so the response isn't delayed
	go sendResetEmail(user.Email, user.FirstName, resetLink)

	return c.Status(200).JSON(fiber.Map{
		"message": "If that email is registered, you will receive a reset link shortly.",
	})
}

// ─── ResetPassword ────────────────────────────────────────────────────────────
// POST /api/reset-password
// Validates the token and sets a new password.
func ResetPassword(c *fiber.Ctx) error {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if strings.TrimSpace(body.Token) == "" || strings.TrimSpace(body.Password) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Token and new password are required"})
	}

	if len(body.Password) < 8 {
		return c.Status(400).JSON(fiber.Map{"error": "Password must be at least 8 characters"})
	}

	// Find user with a valid (non-expired) token
	var user models.User
	if err := database.DB.Where(
		"password_reset_token = ? AND password_reset_expires > ?",
		body.Token,
		time.Now(),
	).First(&user).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Reset link is invalid or has expired"})
	}

	// Hash the new password
	hashed, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to secure password"})
	}

	// Update password and clear the token so it cannot be reused
	if err := database.DB.Model(&user).Updates(map[string]interface{}{
		"password":              string(hashed),
		"password_reset_token":   "",
		"password_reset_expires": nil,
	}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update password"})
	}

	return c.JSON(fiber.Map{"message": "Password reset successfully. You can now log in."})
}

// ─── sendResetEmail ───────────────────────────────────────────────────────────
// Called in a goroutine so it doesn't block the HTTP response.
func sendResetEmail(toEmail, firstName, resetLink string) {
	smtpHost := os.Getenv("SMTP_HOST")      // e.g. smtp.gmail.com
	smtpPort := os.Getenv("SMTP_PORT")      // e.g. 587
	smtpUser := os.Getenv("SMTP_USER")      // sender address
	smtpPass := os.Getenv("SMTP_PASSWORD")  // Gmail App Password or SendGrid key
	fromName := os.Getenv("SMTP_FROM_NAME") // e.g. AfCFTApreneurship Arena

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		// SMTP not configured — log the link so you can test locally
		fmt.Println("[WARN] SMTP not configured. Reset link:")
		fmt.Printf("       %s\n", resetLink)
		return
	}
	if smtpPort == "" {
		smtpPort = "587"
	}
	if fromName == "" {
		fromName = "AfCFTApreneurship Arena"
	}

	subject := "Reset Your Password — " + fromName
	htmlBody := buildResetEmailHTML(firstName, resetLink, fromName)

	msg := []byte(
		"From: " + fromName + " <" + smtpUser + ">\r\n" +
			"To: " + toEmail + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n" +
			"\r\n" +
			htmlBody,
	)

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	if err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{toEmail}, msg); err != nil {
		fmt.Printf("[ERROR] Failed to send reset email to %s: %v\n", toEmail, err)
	} else {
		fmt.Printf("[INFO] Reset email sent to %s\n", toEmail)
	}
}

// ─── buildResetEmailHTML ──────────────────────────────────────────────────────
func buildResetEmailHTML(firstName, resetLink, appName string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8"/>
  <meta name="viewport" content="width=device-width,initial-scale=1.0"/>
  <title>Reset Your Password</title>
</head>
<body style="margin:0;padding:0;background:#f0fdf4;font-family:'Segoe UI',Arial,sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background:#f0fdf4;padding:40px 20px;">
    <tr><td align="center">
      <table width="520" cellpadding="0" cellspacing="0"
        style="background:#fff;border-radius:16px;border:1px solid #d1fae5;overflow:hidden;max-width:100%%;">

        <!-- Header -->
        <tr>
          <td style="background:linear-gradient(135deg,#047857,#10b981);padding:32px 40px;text-align:center;">
            <h1 style="margin:0;color:#fff;font-size:22px;font-weight:700;">%s</h1>
            <p style="margin:6px 0 0;color:rgba(255,255,255,0.75);font-size:13px;">Password Reset Request</p>
          </td>
        </tr>

        <!-- Body -->
        <tr>
          <td style="padding:36px 40px;">
            <p style="margin:0 0 10px;font-size:16px;font-weight:600;color:#0f172a;">Hi %s,</p>
            <p style="margin:0 0 24px;font-size:14px;color:#475569;line-height:1.7;">
              We received a request to reset your password.<br/>
              Click the button below to set a new password.
              This link expires in <strong>1 hour</strong>.
            </p>

            <!-- CTA -->
            <table cellpadding="0" cellspacing="0" style="margin:0 auto 28px;">
              <tr>
                <td style="background:#047857;border-radius:10px;">
                  <a href="%s"
                     style="display:inline-block;padding:14px 32px;color:#fff;font-size:14px;font-weight:700;text-decoration:none;border-radius:10px;">
                    Reset My Password &rarr;
                  </a>
                </td>
              </tr>
            </table>

            <!-- Fallback URL -->
            <p style="margin:0 0 6px;font-size:12px;color:#94a3b8;text-align:center;">
              If the button doesn't work, paste this link into your browser:
            </p>
            <p style="margin:0 0 28px;font-size:11px;color:#047857;word-break:break-all;text-align:center;">%s</p>

            <!-- Warning -->
            <table width="100%%" cellpadding="0" cellspacing="0">
              <tr>
                <td style="background:#fefce8;border:1px solid #fde68a;border-radius:8px;padding:12px 16px;">
                  <p style="margin:0;font-size:12.5px;color:#92400e;line-height:1.6;">
                    ⚠ <strong>Didn't request this?</strong> You can safely ignore this email —
                    your password will not change.
                  </p>
                </td>
              </tr>
            </table>
          </td>
        </tr>

        <!-- Footer -->
        <tr>
          <td style="background:#f8fafc;border-top:1px solid #e2e8f0;padding:20px 40px;text-align:center;">
            <p style="margin:0;font-size:11.5px;color:#94a3b8;">
              &copy; %d %s &middot; AfriTech Systems<br/>
              This is an automated email — please do not reply.
            </p>
          </td>
        </tr>

      </table>
    </td></tr>
  </table>
</body>
</html>`, appName, firstName, resetLink, resetLink, time.Now().Year(), appName)
}