package models

import (
	"time"
)

// Reservation represents a booking made for a hall.
type Reservation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null;size:255" json:"name"`
	Company   string    `gorm:"not null;size:255" json:"company"`
	StartDate time.Time `gorm:"not null" json:"start_date"`
	EndDate   time.Time `gorm:"not null" json:"end_date"`
	TotalCost float64   `gorm:"not null" json:"total_cost"`
	HallID    uint      `gorm:"not null" json:"hall_id"`
	Hall      Hall      `gorm:"foreignKey:HallID" json:"hall,omitempty"`
}

// TableName sets the table name for the Reservation model in the database.
func (Reservation) TableName() string {
	return "reservations"
}

// CalculateTotalCost calculates the total cost of the reservation based on the hall's cost per day.
func (r *Reservation) CalculateTotalCost(costPerDay float64) {
	days := r.EndDate.Sub(r.StartDate).Hours() / 24
	if days < 1 {
		days = 1
	}
	r.TotalCost = days * costPerDay
}
