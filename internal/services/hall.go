package services

import (
	models2 "storage/internal/models"
	. "storage/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"storage/configuration"
	"strconv"
)

// GetHallUtilizationRate calculates the utilization rate of a hall over a specified period.
func GetHallUtilizationRate(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get hall ID from URL parameter.
		hallIDStr := c.Param("id")
		hallID, err := strconv.ParseUint(hallIDStr, 10, 64)

		if err != nil {
			SendError(c, INVALID_HALL_ID, err)
			return
		}

		var hall models2.Hall
		if err := conf.Db.First(&hall, hallID).Error; err != nil {
			SendError(c, HALL_NOT_FOUND, err)
			return
		}

		// Parse optional start_date and end_date query parameters.
		var startDate, endDate time.Time
		if s := c.Query("start_date"); s != "" {
			startDate, err = time.Parse("2006-01-02", s)
			if err != nil {
				SendError(c, INVALID_DATE_FORMAT, err)
				return
			}
		} else {
			// Default: 30 days ago.
			startDate = time.Now().AddDate(0, 0, -30)
		}
		if e := c.Query("end_date"); e != "" {
			endDate, err = time.Parse("2006-01-02", e)
			if err != nil {
				SendError(c, INVALID_DATE_FORMAT, err)
				return
			}
		} else {
			// Default: today.
			endDate = time.Now()
		}

		totalDays := int(endDate.Sub(startDate).Hours()/24) + 1

		// Query reservations for this hall overlapping the period.
		var reservations []models2.Reservation
		if err := conf.Db.
			Where("hall_id = ? AND start_date <= ? AND end_date >= ?", hall.ID, endDate, startDate).
			Find(&reservations).Error; err != nil {
			SendError(c, FAILED_GET_RESERVATIONS, err)
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

		SendSuccessBody(c, gin.H{
			"hall_id":          hall.ID,
			"period":           gin.H{"start": startDate.Format("2006-01-02"), "end": endDate.Format("2006-01-02")},
			"total_days":       totalDays,
			"booked_days":      bookedDays,
			"utilization_rate": utilizationRate,
		})
	}
}
