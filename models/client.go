package models

import (
// "gorm.io/gorm"
)

// Client represents a person using the service
type Client struct {
	ID           uint   `gorm:"primaryKey"`
	FirstName    string `gorm:"not null"`
	LastName     string `gorm:"not null"`
	Email        string `gorm:"unique"`
	PhoneNumber  string
	CompanyName  *string       `gorm:"null"`
	Reservations []Reservation `gorm:"foreignKey:ClientID;constraint:OnDelete:SET NULL"`
}
