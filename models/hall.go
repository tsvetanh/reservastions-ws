package models

// Hall represents a venue that can be reserved.
type Hall struct {
	ID           uint          `gorm:"primaryKey" json:"id"`
	Capacity     int           `gorm:"not null" json:"capacity"`
	Location     string        `gorm:"not null;size:255" json:"location"`
	Available    bool          `gorm:"default:true" json:"available"`
	CostPerDay   float64       `gorm:"not null" json:"cost_per_day"`
	Reservations []Reservation `gorm:"foreignKey:HallID" json:"reservations,omitempty"`
}

// TableName sets the table name for the Hall model in the database.
func (Hall) TableName() string {
	return "halls"
}
