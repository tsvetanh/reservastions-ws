package models

import (
	//"gorm.io/gorm"
	"time"
)

// Employee represents a person working for the hall reservations
type Employee struct {
	ID          uint      `gorm:"primaryKey"`
	FirstName   string    `gorm:"not null;size:100"` //Consistent size limit
	LastName    string    `gorm:"not null;size:100"`
	Email       string    `gorm:"unique;not null;size:150"` //Required email
	PhoneNumber string    `gorm:"size:20"`
	Role        string    `gorm:"not null;size:50"` //Limit to a reasonable role length
	HiredAt     time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}
