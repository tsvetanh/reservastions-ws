package reservation

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"storage/configuration"
	"storage/models"
)

// CreateReservation handles creating a new reservation.
func CreateReservation(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reservation models.Reservation
		if err := c.ShouldBindJSON(&reservation); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		// Ensure valid dates
		if reservation.StartDate.After(reservation.EndDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Start date must be before end date"})
			return
		}

		// Prevent double booking
		var count int64
		conf.Db.Model(&models.Reservation{}).
			Where("hall_id = ? AND ((start_date BETWEEN ? AND ?) OR (end_date BETWEEN ? AND ?))",
				reservation.HallID, reservation.StartDate, reservation.EndDate,
				reservation.StartDate, reservation.EndDate).
			Count(&count)

		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Hall is already booked for these dates"})
			return
		}

		// Get hall price and calculate total cost
		var hall models.Hall
		if err := conf.Db.First(&hall, reservation.HallID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Hall not found"})
			return
		}
		reservation.CalculateTotalCost(hall.CostPerDay)

		// Save the reservation
		if err := conf.Db.Create(&reservation).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reservation"})
			return
		}

		c.JSON(http.StatusOK, reservation)
	}
}

// GetReservations fetches all reservations.
func GetReservations(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reservations []models.Reservation
		if err := conf.Db.Preload("Hall").Find(&reservations).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve reservations"})
			return
		}
		c.JSON(http.StatusOK, reservations)
	}
}

// DeleteReservation removes a reservation.
func DeleteReservation(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := conf.Db.Delete(&models.Reservation{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete reservation"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Reservation deleted successfully"})
	}
}
