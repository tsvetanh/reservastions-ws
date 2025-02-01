package hall

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"storage/configuration"
	"storage/models"
	"storage/services/receipt"
	"time"
)

// CreateHall handles the creation of a new hall.
func CreateHall(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		var hall models.Hall
		if err := c.ShouldBindJSON(&hall); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		// Validate capacity and cost per day.
		if hall.Capacity <= 0 || hall.CostPerDay <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Capacity and cost must be positive numbers"})
			return
		}

		// Validate the available dates, if provided.
		if !hall.AvailableFrom.IsZero() && !hall.AvailableTo.IsZero() {
			if !hall.AvailableFrom.Before(hall.AvailableTo) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "AvailableFrom must be before AvailableTo"})
				return
			}
			// Ensure the AvailableFrom date is not in the past.
			if hall.AvailableFrom.Before(time.Now()) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "AvailableFrom cannot be in the past"})
				return
			}
		}

		// Save the hall in the database.
		if err := conf.Db.Create(&hall).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create hall"})
			return
		}

		c.JSON(http.StatusOK, hall)
	}
}

// CreateReservation handles creating a new reservation.
func CreateReservation(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reservation models.Reservation
		if err := c.ShouldBindJSON(&reservation); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		// Ensure the start date is before the end date.
		if !reservation.StartDate.Before(reservation.EndDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Start date must be before end date"})
			return
		}

		// Ensure the start date is not in the past.
		if reservation.StartDate.Before(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Start date cannot be in the past"})
			return
		}

		// Prevent double booking.
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

		// Fetch hall price and calculate total cost.
		var hall models.Hall
		if err := conf.Db.First(&hall, reservation.HallID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Hall not found"})
			return
		}
		reservation.CalculateTotalCost(hall.CostPerDay)

		// Save the reservation.
		if err := conf.Db.Create(&reservation).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reservation"})
			return
		}

		// Generate receipt after successful creation.
		if err := receipt.GenerateReceipt(&reservation); err != nil {
			// Optionally log the error or notify the admin; the reservation creation succeeded.
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Reservation created, but failed to generate receipt"})
			return
		}

		// (Optional) Compute additional details such as duration, cost per day, etc.
		duration := int(reservation.EndDate.Sub(reservation.StartDate).Hours() / 24)
		if duration < 1 {
			duration = 1
		}
		c.JSON(http.StatusOK, gin.H{
			"reservation": reservation,
			"details": gin.H{
				"duration_days": duration,
				"cost_per_day":  reservation.TotalCost / float64(duration),
			},
		})
	}
}

// GetHalls retrieves all available halls.
func GetHalls(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		var halls []models.Hall
		if err := conf.Db.Find(&halls).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve halls"})
			return
		}
		c.JSON(http.StatusOK, halls)
	}
}

// UpdateHall modifies an existing hall.
func UpdateHall(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var hall models.Hall
		if err := conf.Db.First(&hall, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Hall not found"})
			return
		}

		if err := c.ShouldBindJSON(&hall); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		conf.Db.Save(&hall)
		c.JSON(http.StatusOK, hall)
	}
}

// DeleteHall removes a hall.
func DeleteHall(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := conf.Db.Delete(&models.Hall{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete hall"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Hall deleted successfully"})
	}
}
