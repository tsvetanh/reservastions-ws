package models

import (
// "gorm.io/gorm"
)

// Client represents a person using the service
type Client struct {
	ID           uint          `gorm:"primaryKey"`
	FirstName    string        `gorm:"not null;size:100"`        // Set a reasonable max length
	LastName     string        `gorm:"not null;size:100"`        //Set a reasonable max length
	Email        string        `gorm:"unique;not null;size:150"` //Ensure it's required
	PhoneNumber  string        `gorm:"size:20"`                  //Restrict size
	CompanyName  *string       `gorm:"size:150"`                 //Prevent excessively long names
	Reservations []Reservation `gorm:"foreignKey:ClientID;constraint:OnDelete:SET NULL"`
}
