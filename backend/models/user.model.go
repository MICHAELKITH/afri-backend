package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a registered user
type User struct {
	gorm.Model // includes ID, CreatedAt, UpdatedAt, DeletedAt

	FirstName     string `gorm:"not null"        json:"first_name"`
	LastName      string `gorm:"not null"        json:"last_name"`
	Email         string `gorm:"uniqueIndex;not null" json:"email"`
	PhoneNumber   string `gorm:"uniqueIndex"     json:"phone_number"`
	Country       string `gorm:"not null"        json:"country"`
	StudyLevel    string `gorm:"not null"        json:"study_level"`
	FieldOfStudy  string `gorm:"not null"        json:"field_of_study"`
	YearOfStudy   int    `                       json:"year_of_study"`
	LearningGoals string `                       json:"learning_goals"`
	Password      string `gorm:"not null"        json:"password"`

	// Password reset — populated only during forgot-password flow
	PasswordResetToken   string     `gorm:"column:password_reset_token"   json:"-"`
	PasswordResetExpires *time.Time `gorm:"column:password_reset_expires" json:"-"`
}

// Safe returns user data without sensitive fields
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
		"created_at":     u.CreatedAt,
		"updated_at":     u.UpdatedAt,
	}
}