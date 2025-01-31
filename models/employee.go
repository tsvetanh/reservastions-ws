package models

import (
	//"gorm.io/gorm"
	"time"
)

// Employee represents a person working for the hall reservations
type Employee struct {
	ID          uint   `gorm:"primaryKey"`
	FirstName   string `gorm:"not null"`
	LastName    string `gorm:"not null"`
	Email       string `gorm:"unique"`
	PhoneNumber string
	Role        string    `gorm:"not null"`
	HiredAt     time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}
