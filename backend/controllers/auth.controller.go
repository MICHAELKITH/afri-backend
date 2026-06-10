package controllers

import (
	"backend/database"
	"backend/models"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ─── Helpers ──────────────────────────────────────────────────────────────────

func getJWTSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

func generateOTP() (string, error) {
	max := big.NewInt(1_000_000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// issueToken creates a signed JWT for the given user (includes role).
func issueToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
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

	// Validate role — only allow known roles
	req.Role = strings.TrimSpace(strings.ToLower(req.Role))
	if req.Role != "student" && req.Role != "trader" {
		req.Role = "student" // default
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
		"role":           req.Role,
	}

	if pn := strings.TrimSpace(req.PhoneNumber); pn != "" {
		createData["phone_number"] = pn
	}

if err := database.DB.Model(&models.User{}).Create(createData).Error; err != nil {
    if strings.Contains(err.Error(), "23505") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
        // Determine which field caused the conflict
        field := "unknown"
        errStr := err.Error()
        switch {
        case strings.Contains(errStr, "idx_users_email") || strings.Contains(errStr, "uni_users_email"):
            field = "email"
        case strings.Contains(errStr, "idx_users_phone_number") || strings.Contains(errStr, "uni_users_phone_number"):
            field = "phone_number"
        }

        return c.Status(400).JSON(fiber.Map{
            "error":   "Duplicate value",
            "field":   field,
            "message": "This " + field + " is already registered",
        })
    }
    return c.Status(400).JSON(fiber.Map{"error": "Registration failed: invalid data provided"})
}

	if err := database.DB.Where("email = ?", req.Email).First(&req).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch created user"})
	}

	signed, err := issueToken(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error: token generation failed"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Welcome, " + req.FirstName + "!",
		"token":   signed,
		"user":    req.Safe(),
	})
}

func CheckEmail(c *fiber.Ctx) error {
    var body struct {
        Email string `json:"email"`
    }
    if err := c.BodyParser(&body); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }

    body.Email = strings.TrimSpace(strings.ToLower(body.Email))
    if body.Email == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Email is required"})
    }

    var user models.User
    if err := database.DB.Where("email = ?", body.Email).First(&user).Error; err == nil {
        // Email exists
        return c.Status(409).JSON(fiber.Map{
            "error": "Email already registered",
            "field": "email",
        })
    }

    // Email available
    return c.Status(200).JSON(fiber.Map{
        "message": "Email available",
        "field":   "email",
    })
}

// ─── Login (generic — kept for backwards compatibility) ───────────────────────

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

	signed, err := issueToken(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Login failed: could not issue token"})
	}

	return c.JSON(fiber.Map{"token": signed, "user": user.Safe()})
}

// ─── Role-specific Login ──────────────────────────────────────────────────────

func LoginStudent(c *fiber.Ctx) error {
	return loginByRole(c, "student")
}

func LoginTrader(c *fiber.Ctx) error {
	return loginByRole(c, "trader")
}

// ─── Role-specific Login ──────────────────────────────────────────────────────

func loginByRole(c *fiber.Ctx, role string) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request format"})
	}

	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	if body.Email == "" || body.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email and password are required"})
	}

	// Step 1: find by email only — no role filter
	var user models.User
	if err := database.DB.Where("email = ?", body.Email).First(&user).Error; err != nil {
		// Email doesn't exist — give a vague message
		return c.Status(401).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	// Step 2: verify password before revealing anything about the account
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	// Step 3: password is confirmed — now safe to check role
	if user.Role != role {
		return c.Status(403).JSON(fiber.Map{
			"error": fmt.Sprintf("This account is registered as a %s. Please use the %s login.", user.Role, user.Role),
			"role":  user.Role,
		})
	}

	signed, err := issueToken(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Login failed: could not issue token"})
	}

	return c.JSON(fiber.Map{"token": signed, "user": user.Safe()})
}

// ─── Logout ───────────────────────────────────────────────────────────────────

func Logout(c *fiber.Ctx) error {
	c.ClearCookie("afridauth")
	return c.JSON(fiber.Map{"message": "Successfully logged out"})
}

// ─── ForgotPassword ───────────────────────────────────────────────────────────

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

	genericMsg := fiber.Map{"message": "If that email is registered, you will receive a reset code shortly."}

	var user models.User
	if err := database.DB.Where("email = ?", body.Email).First(&user).Error; err != nil {
		return c.Status(200).JSON(genericMsg)
	}

	otp, err := generateOTP()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate OTP"})
	}

	hashedOTP, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to secure OTP"})
	}

	expiresAt := time.Now().Add(10 * time.Minute)

	if err := database.DB.Model(&user).Updates(map[string]interface{}{
		"password_reset_token":   string(hashedOTP),
		"password_reset_expires": expiresAt,
	}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save OTP"})
	}

	go sendOTPEmail(user.Email, user.FirstName, otp)

	return c.Status(200).JSON(genericMsg)
}

// ─── ResetPassword ────────────────────────────────────────────────────────────

func ResetPassword(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		OTP      string `json:"otp"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	body.OTP = strings.TrimSpace(body.OTP)

	if body.Email == "" || body.OTP == "" || body.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email, OTP, and new password are required"})
	}

	if len(body.Password) < 8 {
		return c.Status(400).JSON(fiber.Map{"error": "Password must be at least 8 characters"})
	}

	var user models.User
	if err := database.DB.Where(
		"email = ? AND password_reset_expires > ?",
		body.Email,
		time.Now(),
	).First(&user).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "OTP is invalid or has expired"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordResetToken), []byte(body.OTP)); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Incorrect OTP. Please check your email and try again."})
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to secure password"})
	}

	if err := database.DB.Model(&user).Updates(map[string]interface{}{
		"password":               string(hashed),
		"password_reset_token":   "",
		"password_reset_expires": nil,
	}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update password"})
	}

	return c.JSON(fiber.Map{"message": "Password reset successfully. You can now log in."})
}

// ─── sendOTPEmail ─────────────────────────────────────────────────────────────

func sendOTPEmail(toEmail, firstName, otp string) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASSWORD")
	fromName := os.Getenv("SMTP_FROM_NAME")

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		fmt.Println("[WARN] SMTP not configured. OTP:")
		fmt.Printf("       %s\n", otp)
		return
	}
	if smtpPort == "" {
		smtpPort = "587"
	}
	if fromName == "" {
		fromName = "AfCFTApreneurship Arena"
	}

	subject := "Your Password Reset Code — " + fromName
	htmlBody := buildOTPEmailHTML(firstName, otp, fromName)

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
		fmt.Printf("[ERROR] Failed to send OTP email to %s: %v\n", toEmail, err)
	} else {
		fmt.Printf("[INFO] OTP email sent to %s\n", toEmail)
	}
}

// ─── buildOTPEmailHTML ────────────────────────────────────────────────────────

func buildOTPEmailHTML(firstName, otp, appName string) string {
	digits := strings.Split(otp, "")
	var digitBoxes string
	for _, d := range digits {
		digitBoxes += fmt.Sprintf(`
		  <td style="padding:0 4px;">
		    <div style="
		      width:44px;height:54px;
		      border:2px solid #d1fae5;
		      border-radius:10px;
		      background:#f0fdf4;
		      display:flex;align-items:center;justify-content:center;
		      font-size:28px;font-weight:700;color:#047857;
		      font-family:'Segoe UI',Arial,sans-serif;
		      line-height:54px;text-align:center;
		    ">%s</div>
		  </td>`, d)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8"/>
  <meta name="viewport" content="width=device-width,initial-scale=1.0"/>
  <title>Your Reset Code</title>
</head>
<body style="margin:0;padding:0;background:#f0fdf4;font-family:'Segoe UI',Arial,sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background:#f0fdf4;padding:40px 20px;">
    <tr><td align="center">
      <table width="520" cellpadding="0" cellspacing="0"
        style="background:#fff;border-radius:16px;border:1px solid #d1fae5;overflow:hidden;max-width:100%%;">
        <tr>
          <td style="background:linear-gradient(135deg,#047857,#10b981);padding:32px 40px;text-align:center;">
            <h1 style="margin:0;color:#fff;font-size:22px;font-weight:700;">%s</h1>
            <p style="margin:6px 0 0;color:rgba(255,255,255,0.75);font-size:13px;">Password Reset Code</p>
          </td>
        </tr>
        <tr>
          <td style="padding:36px 40px;">
            <p style="margin:0 0 10px;font-size:16px;font-weight:600;color:#0f172a;">Hi %s,</p>
            <p style="margin:0 0 28px;font-size:14px;color:#475569;line-height:1.7;">
              Use the 6-digit code below to reset your password.
              This code expires in <strong>10 minutes</strong> and can only be used once.
            </p>
            <table cellpadding="0" cellspacing="0" style="margin:0 auto 28px;">
              <tr>%s</tr>
            </table>
            <p style="margin:0 0 28px;font-size:12px;color:#94a3b8;text-align:center;">
              Your code: <strong style="color:#047857;letter-spacing:4px;">%s</strong>
            </p>
            <table width="100%%" cellpadding="0" cellspacing="0">
              <tr>
                <td style="background:#fefce8;border:1px solid #fde68a;border-radius:8px;padding:12px 16px;">
                  <p style="margin:0;font-size:12.5px;color:#92400e;line-height:1.6;">
                    ⚠ <strong>Didn't request this?</strong> You can safely ignore this email —
                    your password will not change. Never share this code with anyone.
                  </p>
                </td>
              </tr>
            </table>
          </td>
        </tr>
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
</html>`, appName, firstName, digitBoxes, otp, time.Now().Year(), appName)
}