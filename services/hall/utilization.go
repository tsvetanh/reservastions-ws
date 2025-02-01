package hall

import (
	"net/http"
	"reservations/services/reservation"
	"time"

	"github.com/gin-gonic/gin"
	"reservations/configuration"
	"strconv"
)

// GetHallUtilizationRate calculates the utilization rate of a hall over a specified period.
func GetHallUtilizationRate(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get hall ID from URL parameter.
		hallIDStr := c.Param("id")
		hallID, err := strconv.ParseUint(hallIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hall ID"})
			return
		}

		var hall Hall
		if err := conf.Db.First(&hall, hallID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Hall not found"})
			return
		}

		// Parse optional start_date and end_date query parameters.
		var startDate, endDate time.Time
		if s := c.Query("start_date"); s != "" {
			startDate, err = time.Parse("2006-01-02", s)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
				return
			}
		} else {
			// Default: 30 days ago.
			startDate = time.Now().AddDate(0, 0, -30)
		}
		if e := c.Query("end_date"); e != "" {
			endDate, err = time.Parse("2006-01-02", e)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
				return
			}
		} else {
			// Default: today.
			endDate = time.Now()
		}

		totalDays := int(endDate.Sub(startDate).Hours()/24) + 1

		// Query reservations for this hall overlapping the period.
		var reservations []reservation.Reservation
		if err := conf.Db.
			Where("hall_id = ? AND start_date <= ? AND end_date >= ?", hall.ID, endDate, startDate).
			Find(&reservations).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve reservations"})
			return
		}

		// Compute booked days by summing overlaps.
		bookedDays := 0
		for _, r := range reservations {
			overlapStart := r.StartDate
			if startDate.After(overlapStart) {
				overlapStart = startDate
			}
			overlapEnd := r.EndDate
			if endDate.Before(overlapEnd) {
				overlapEnd = endDate
			}
			days := int(overlapEnd.Sub(overlapStart).Hours()/24) + 1
			bookedDays += days
		}

		utilizationRate := (float64(bookedDays) / float64(totalDays)) * 100

		c.JSON(http.StatusOK, gin.H{
			"hall_id":          hall.ID,
			"period":           gin.H{"start": startDate.Format("2006-01-02"), "end": endDate.Format("2006-01-02")},
			"total_days":       totalDays,
			"booked_days":      bookedDays,
			"utilization_rate": utilizationRate,
		})
	}
}
