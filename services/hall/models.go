package hall

import (
	"reservations/services/reservation"
	"time"
)

// Hall represents a venue that can be reserved.
type Hall struct {
	ID         uint    `gorm:"primaryKey" json:"id"`
	Capacity   int     `gorm:"not null" json:"capacity"`
	Location   string  `gorm:"not null;size:255" json:"location"`
	Available  bool    `gorm:"default:true" json:"available"`
	CostPerDay float64 `gorm:"not null" json:"cost_per_day"`
	// New fields for available dates:
	AvailableFrom time.Time                 `json:"available_from"`
	AvailableTo   time.Time                 `json:"available_to"`
	Reservations  []reservation.Reservation `gorm:"foreignKey:HallID" json:"reservations,omitempty"`
	HallImages    []HallImage               `gorm:"foreignKey:HallID" json:"-"` // Prevent images from being included directly in the response
	ImageURLs     []string                  `json:"images"`                     // Add this field to store image URLs
}

type HallImage struct {
	ID       uint   `gorm:"primaryKey"`
	HallID   uint   `gorm:"not null"`
	ImageURL string `gorm:"size:255"`
	Hall     Hall   `gorm:"foreignKey:HallID"`
}

// TableName sets the table name for the Hall model in the database.
func (Hall) TableName() string {
	return "hall_res_project.halls"
}

func (HallImage) TableName() string {
	return "hall_res_project.halls_images"
}
