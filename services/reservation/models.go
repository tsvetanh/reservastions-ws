package reservation

import (
	"reservations/services/hall"
	"time"
)

type Reservation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null;size:255" json:"name"`
	Company   string    `gorm:"not null;size:255" json:"company"`
	StartDate time.Time `gorm:"not null" json:"start_date"`
	EndDate   time.Time `gorm:"not null" json:"end_date"`
	TotalCost float64   `gorm:"not null" json:"total_cost"`
	HallID    uint      `gorm:"not null" json:"hall_id"`
	Hall      hall.Hall `gorm:"foreignKey:HallID" json:"hall,omitempty"`
}

// TableName sets the table name for the Reservation model in the database.
func (Reservation) TableName() string {
	return "hall_res_project.reservations"
}

// CalculateTotalCost calculates the total cost of the reservation based on the hall's cost per day.
// If the reservation lasts longer than 7 days, a 10% discount is applied.
func (r *Reservation) CalculateTotalCost(costPerDay float64) {
	// Calculate the number of days.
	days := r.EndDate.Sub(r.StartDate).Hours() / 24
	if days < 1 {
		days = 1
	}

	// Calculate total cost without discount.
	total := days * costPerDay

	// Apply a 10% discount for reservations longer than 7 days.
	if days > 7 {
		discount := total * 0.10
		total -= discount
	}

	r.TotalCost = total
}
