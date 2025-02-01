// services/reservation/summary.go
package reservation

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"storage/configuration"
	"storage/models"
)

// GetReservationSummary aggregates reservation data for dashboard display.
func GetReservationSummary(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reservations []models.Reservation

		//Skip DB operations if DB is not initialized
		if conf.Db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database is disabled. Cannot create reservation."})
			return
		}

		if err := conf.Db.Find(&reservations).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve reservations"})
			return
		}

		now := time.Now()
		var pastCount, currentCount, upcomingCount int
		var totalRevenue float64

		for _, r := range reservations {
			totalRevenue += r.TotalCost
			if r.EndDate.Before(now) {
				pastCount++
			} else if r.StartDate.After(now) {
				upcomingCount++
			} else {
				currentCount++
			}
		}

		summary := gin.H{
			"total_reservations":    len(reservations),
			"past_reservations":     pastCount,
			"current_reservations":  currentCount,
			"upcoming_reservations": upcomingCount,
			"total_revenue":         totalRevenue,
		}
		c.JSON(http.StatusOK, summary)
	}
}
