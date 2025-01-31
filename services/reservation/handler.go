package reservation

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"storage/configuration"
	"storage/models"
	"time"
)

// UpdateReservation modifies an existing reservation
func UpdateReservation(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Fetch the existing reservation
		var reservation models.Reservation
		if err := conf.Db.First(&reservation, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reservation not found"})
			return
		}

		// Bind the incoming JSON to the reservation struct
		var updatedReservation models.Reservation
		if err := c.ShouldBindJSON(&updatedReservation); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		// Ensure start_date < end_date
		if !updatedReservation.StartDate.Before(updatedReservation.EndDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Start date must be before end date"})
			return
		}

		// Check for overlapping reservations (prevent double booking)
		var count int64
		conf.Db.Model(&models.Reservation{}).
			Where("hall_id = ? AND id != ? AND ((start_date BETWEEN ? AND ?) OR (end_date BETWEEN ? AND ?))",
				updatedReservation.HallID, id,
				updatedReservation.StartDate, updatedReservation.EndDate,
				updatedReservation.StartDate, updatedReservation.EndDate).
			Count(&count)

		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "The hall is already booked for the selected dates"})
			return
		}

		// Fetch the hall's cost per day
		var hall models.Hall
		if err := conf.Db.First(&hall, updatedReservation.HallID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Hall not found"})
			return
		}

		// Update reservation fields
		reservation.Name = updatedReservation.Name
		reservation.Company = updatedReservation.Company
		reservation.HallID = updatedReservation.HallID
		reservation.StartDate = updatedReservation.StartDate
		reservation.EndDate = updatedReservation.EndDate

		// Calculate the updated total cost using the hall's price per day
		reservation.CalculateTotalCost(hall.CostPerDay)

		// Save the updated reservation
		if err := conf.Db.Save(&reservation).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update reservation"})
			return
		}

		c.JSON(http.StatusOK, reservation)
	}
}

// CreateReservation handles creating a new reservation
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

// GetReservations retrieves reservations with optional filtering
func GetReservations(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reservations []models.Reservation
		query := conf.Db

		// Apply filters if query parameters are provided
		if date := c.Query("date"); date != "" {
			parsedDate, err := time.Parse("2006-01-02", date)
			if err == nil {
				query = query.Where("start_date <= ? AND end_date >= ?", parsedDate, parsedDate)
			}
		}

		if company := c.Query("company"); company != "" {
			query = query.Where("company = ?", company)
		}

		if hallID := c.Query("hall"); hallID != "" {
			query = query.Where("hall_id = ?", hallID)
		}

		// Fetch data
		if err := query.Find(&reservations).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve reservations"})
			return
		}

		c.JSON(http.StatusOK, reservations)
	}
}

// DeleteReservation removes a reservation and its receipt file
func DeleteReservation(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Fetch the reservation to check if it exists
		var reservation models.Reservation
		if err := conf.Db.First(&reservation, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reservation not found"})
			return
		}

		// Delete the reservation from the database
		if err := conf.Db.Delete(&reservation).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete reservation"})
			return
		}

		// Try deleting the associated receipt file
		if err := deleteReceiptFile(reservation.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Reservation deleted, but failed to delete receipt"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Reservation and receipt deleted successfully"})
	}
}

// deleteReceiptFile removes the receipt file associated with a reservation
func deleteReceiptFile(reservationID uint) error {
	receiptDir := "receipts"
	filename := fmt.Sprintf("receipt_%d.txt", reservationID)
	filePath := filepath.Join(receiptDir, filename)

	// Check if the file exists before trying to delete it
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // If file doesn't exist, there's nothing to delete
	}

	// Remove the file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete receipt file: %v", err)
	}

	fmt.Println("Receipt file deleted:", filePath)
	return nil
}
