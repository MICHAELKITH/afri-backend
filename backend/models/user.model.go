package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey;autoIncrement"     json:"id"`
	CreatedAt time.Time      `                                    json:"created_at"`
	UpdatedAt time.Time      `                                    json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                        json:"-"`

	FirstName     string `gorm:"not null;size:100"             json:"first_name"`
	LastName      string `gorm:"not null;size:100"             json:"last_name"`
	Email         string `gorm:"uniqueIndex;not null;size:150" json:"email"`
	PhoneNumber   string `gorm:"uniqueIndex;size:20"           json:"phone_number"`
	Country       string `gorm:"not null;size:100"             json:"country"`
	StudyLevel    string `gorm:"not null;size:100"             json:"study_level"`
	FieldOfStudy  string `gorm:"not null;size:150"             json:"field_of_study"`
	YearOfStudy   int    `                                     json:"year_of_study"`
	LearningGoals string `gorm:"type:text"                    json:"learning_goals"`
	Password      string `gorm:"not null;size:255"             json:"password"`
	Role          string `gorm:"not null;default:student;size:50" json:"role"`

	// ── Email verification ────────────────────────────────────────────────────
	EmailVerified      bool       `gorm:"column:email_verified;default:false"        json:"email_verified"`
	EmailVerifyToken   string     `gorm:"column:email_verify_token;size:255"         json:"-"`
	EmailVerifyExpires *time.Time `gorm:"column:email_verify_expires"               json:"-"`

	// ── Password reset ────────────────────────────────────────────────────────
	PasswordResetToken   string     `gorm:"column:password_reset_token;size:255"  json:"-"`
	PasswordResetExpires *time.Time `gorm:"column:password_reset_expires"         json:"-"`
}

func (u *User) Safe() map[string]interface{} {
	return map[string]interface{}{
		"id":             u.ID,
		"first_name":     u.FirstName,
		"last_name":      u.LastName,
		"email":          u.Email,
		"phone_number":   u.PhoneNumber,
		"country":        u.Country,
		"study_level":    u.StudyLevel,
		"field_of_study": u.FieldOfStudy,
		"year_of_study":  u.YearOfStudy,
		"learning_goals": u.LearningGoals,
		"role":           u.Role,
		"email_verified": u.EmailVerified,
		"created_at":     u.CreatedAt,
		"updated_at":     u.UpdatedAt,
	}
}