package models

import "time"

// DateRange represents a suggested available period.
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
