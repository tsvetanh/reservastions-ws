package models

import "time"

// Hall represents a venue that can be reserved.
type Hall struct {
	ID         uint    `gorm:"primaryKey" json:"id"`
	Name       string  `gorm:"not null" json:"name"`
	Capacity   int     `gorm:"not null" json:"capacity"`
	Location   string  `gorm:"not null;size:255" json:"location"`
	Available  bool    `gorm:"default:true" json:"available"`
	CostPerDay float64 `gorm:"not null" json:"cost_per_day"`
	Category   string  `gorm:"category" json:"category"`
	// New fields for available dates:
	AvailableFrom time.Time     `gorm:"default:CURRENT_TIMESTAMP(3)" json:"available_from"`
	AvailableTo   time.Time     `gorm:"default:CURRENT_TIMESTAMP(3)" json:"available_to"`
	Reservations  []Reservation `gorm:"foreignKey:HallID" json:"reservations,omitempty"`
	HallImages    []HallImage   `gorm:"foreignKey:HallID" json:"-"`
	ImageURLs     []string      `gorm:"-" json:"images"`
}

type HallImage struct {
	ID        uint   `gorm:"primaryKey"`
	HallID    uint   `gorm:"not null"`
	ImageName string `gorm:"size:255"`
	Hall      Hall   `gorm:"foreignKey:HallID"`
}

// TableName sets the table name for the Hall model in the database.
func (Hall) TableName() string {
	return "hall_res_project.halls"
}

func (HallImage) TableName() string {
	return "hall_res_project.halls_images"
}
