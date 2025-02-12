package reservation

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"storage/configuration"
	"storage/models"
	"storage/services/receipt"
	"strings"
	"time"
)

// CreateReservation handles creating a new reservation.
func CreateReservation(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reservation models.Reservation

		// Extract authenticated user ID (from middleware)
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		reservation.UserID = userID.(int64) // Store the UserID

		// Skip DB operations if DB is not initialized
		if conf.Db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database is disabled. Cannot create reservation."})
			return
		}

		// Parse request body into reservation model
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

		// Prevent double booking by checking overlapping reservations.
		var count int64
		conf.Db.Model(&models.Reservation{}).
			Where("hall_id = ? AND ((start_date BETWEEN ? AND ?) OR (end_date BETWEEN ? AND ?))",
				reservation.HallID, reservation.StartDate, reservation.EndDate,
				reservation.StartDate, reservation.EndDate).
			Count(&count)

		// Suggest alternative dates if the hall is already booked
		if count > 0 {
			suggestions, err := SuggestAlternativeDates(conf, reservation.HallID, reservation.StartDate, reservation.EndDate)
			if err != nil {
				c.JSON(http.StatusConflict, gin.H{"error": "Hall is already booked for these dates"})
			} else {
				c.JSON(http.StatusConflict, gin.H{
					"error":       "Hall is already booked for these dates",
					"suggestions": suggestions,
				})
			}
			return
		}

		// Fetch hall price to calculate total cost.
		var hall models.Hall
		if err := conf.Db.First(&hall, reservation.HallID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Hall not found"})
			return
		}
		reservation.CalculateTotalCost(hall.CostPerDay)

		// Save the reservation in the database.
		if err := conf.Db.Create(&reservation).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reservation"})
			return
		}

		// Generate a receipt after successful reservation creation.
		if err := receipt.GenerateReceipt(&reservation); err != nil {
			// Log error but still return success since reservation was created.
			fmt.Printf("Warning: Failed to generate receipt for reservation ID %d: %v\n", reservation.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Reservation created, but failed to generate receipt"})
			return
		}

		// Compute reservation duration and cost per day.
		duration := int(reservation.EndDate.Sub(reservation.StartDate).Hours() / 24)
		if duration < 1 {
			duration = 1
		}

		// Return success response with reservation details.
		c.JSON(http.StatusOK, gin.H{
			"reservation": reservation,
			"details": gin.H{
				"duration_days": duration,
				"cost_per_day":  reservation.TotalCost / float64(duration),
			},
		})
	}
}

// UpdateReservation modifies an existing reservation
func UpdateReservation(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Fetch the existing reservation
		var reservation models.Reservation

		//Skip DB operations if DB is not initialized
		if conf.Db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database is disabled. Cannot create reservation."})
			return
		}

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

// GetReservations retrieves reservations with filtering and sorting, ensuring users see only their reservations unless they are admins.
func GetReservations(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reservations []models.Reservation
		query := conf.Db

		// Skip DB operations if DB is not initialized
		if conf.Db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database is disabled. Cannot retrieve reservations."})
			return
		}

		// Extract user ID and roles from authentication
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		roles, _ := c.Get("roles") // Get user roles from middleware
		isAdmin := false

		// Check if user has admin role
		if rolesList, ok := roles.([]string); ok {
			for _, role := range rolesList {
				if strings.ToLower(role) == "admin" {
					isAdmin = true
					break
				}
			}
		}

		// Restrict non-admin users to their own reservations
		if !isAdmin {
			query = query.Where("user_id = ?", userID)
		}

		// Filter by a specific date (Format: "YYYY-MM-DD")
		if dateStr := c.Query("date"); dateStr != "" {
			parsedDate, err := time.Parse("2006-01-02", dateStr)
			if err == nil {
				query = query.Where("start_date <= ? AND end_date >= ?", parsedDate, parsedDate)
			}
		}

		// Filter by company (case-insensitive)
		if company := c.Query("company"); company != "" {
			query = query.Where("LOWER(company) = ?", strings.ToLower(company))
		}

		// Filter by hall (hall ID)
		if hall := c.Query("hall"); hall != "" {
			query = query.Where("hall_id = ?", hall)
		}

		// Sorting: "sort_by" and "order" parameters (Allowed: start_date, end_date, company, hall_id)
		if sortBy := c.Query("sort_by"); sortBy != "" {
			order := c.DefaultQuery("order", "asc")
			if order != "asc" && order != "desc" {
				order = "asc"
			}

			allowedSortFields := map[string]bool{
				"start_date": true,
				"end_date":   true,
				"company":    true,
				"hall_id":    true,
			}

			if allowedSortFields[sortBy] {
				query = query.Order(fmt.Sprintf("%s %s", sortBy, order))
			}
		}

		// Execute the query
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

		//Skip DB operations if DB is not initialized
		if conf.Db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database is disabled. Cannot create reservation."})
			return
		}

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

// GetCategorizedReservations groups reservations into Past, Current, and Upcoming.
func GetCategorizedReservations(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reservations []models.Reservation

		//Skip DB operations if DB is not initialized
		if conf.Db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database is disabled. Cannot create reservation."})
			return
		}

		// Preload the Hall association if you need hall details in the response.
		if err := conf.Db.Preload("Hall").Find(&reservations).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve reservations"})
			return
		}

		now := time.Now()
		var past, current, upcoming []models.Reservation

		for _, r := range reservations {
			// Categorize based on the current time relative to reservation dates.
			if r.EndDate.Before(now) {
				past = append(past, r)
			} else if r.StartDate.After(now) {
				upcoming = append(upcoming, r)
			} else {
				current = append(current, r)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"past":     past,
			"current":  current,
			"upcoming": upcoming,
		})
	}
}
