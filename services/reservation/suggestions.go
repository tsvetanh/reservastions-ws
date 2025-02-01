package reservation

import (
	"reservations/configuration"
	"reservations/models"
	"time"
)

// SuggestAlternativeDates queries reservations for a given hall and returns available date ranges
// that can accommodate a reservation of the same duration as the requested one.
func SuggestAlternativeDates(conf *configuration.Dependencies, hallID uint, requestedStart, requestedEnd time.Time) ([]models.DateRange, error) {
	// Define a wider window to search for available gaps (e.g., 30 days before/after the requested dates)
	startWindow := requestedStart.AddDate(0, 0, -30)
	endWindow := requestedEnd.AddDate(0, 0, 30)

	var reservations []Reservation
	if err := conf.Db.
		Where("hall_id = ? AND start_date >= ? AND end_date <= ?", hallID, startWindow, endWindow).
		Order("start_date asc").
		Find(&reservations).Error; err != nil {
		return nil, err
	}

	var suggestions []models.DateRange
	requestedDuration := requestedEnd.Sub(requestedStart)

	// If no reservations exist, suggest the requested dates as available.
	if len(reservations) == 0 {
		suggestions = append(suggestions, models.DateRange{Start: requestedStart, End: requestedStart.Add(requestedDuration)})
		return suggestions, nil
	}

	// Check gap before the first reservation.
	if reservations[0].StartDate.Sub(startWindow) >= requestedDuration {
		suggestions = append(suggestions, models.DateRange{Start: startWindow, End: startWindow.Add(requestedDuration)})
	}

	// Check gaps between reservations.
	for i := 0; i < len(reservations)-1; i++ {
		gap := reservations[i+1].StartDate.Sub(reservations[i].EndDate)
		if gap >= requestedDuration {
			suggestions = append(suggestions, models.DateRange{
				Start: reservations[i].EndDate,
				End:   reservations[i].EndDate.Add(requestedDuration),
			})
		}
	}

	// Check gap after the last reservation.
	if endWindow.Sub(reservations[len(reservations)-1].EndDate) >= requestedDuration {
		suggestions = append(suggestions, models.DateRange{
			Start: reservations[len(reservations)-1].EndDate,
			End:   reservations[len(reservations)-1].EndDate.Add(requestedDuration),
		})
	}

	return suggestions, nil
}
