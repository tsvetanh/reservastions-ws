package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"storage/configuration"
	"storage/internal/models"
	"storage/internal/receipt"
	"storage/internal/repos"
	. "storage/internal/services"
	. "storage/internal/utils"
	"strings"
	"time"
)

type ReservationHandler struct {
	Repo *repos.BaseRepository[models.Reservation]
	Conf *configuration.Dependencies
}

func NewReservationHandler(conf *configuration.Dependencies) *ReservationHandler {
	return &ReservationHandler{
		Repo: &repos.BaseRepository[models.Reservation]{DB: conf.Db},
		Conf: conf,
	}
}

// CreateReservation handles creating a new reservation.
func (h *ReservationHandler) CreateReservation() gin.HandlerFunc {
	return func(c *gin.Context) {
		var reservation models.Reservation

		// Validate input
		if err := c.ShouldBindJSON(&reservation); err != nil {
			SendError(c, INVALID_REQ_PAYLOAD, err)
			return
		}

		if !reservation.StartDate.Before(reservation.EndDate) || reservation.StartDate.Before(time.Now()) {
			SendError(c, INVALID_RES_START_DATE, nil)
			return
		}

		// Check for conflicts
		var count int64
		h.Conf.Db.Model(&models.Reservation{}).
			Where("hall_id = ? AND ((start_date BETWEEN ? AND ?) OR (end_date BETWEEN ? AND ?))",
				reservation.HallID, reservation.StartDate, reservation.EndDate,
				reservation.StartDate, reservation.EndDate).
			Count(&count)

		if count > 0 {
			suggestions, err := SuggestAlternativeDates(h.Conf, reservation.HallID, reservation.StartDate, reservation.EndDate)
			if err != nil {
				SendError(c, HALL_BOOKED, err)
			} else {
				SendErrorBody(c, HALL_BOOKED, gin.H{"suggestions": suggestions}, nil)
			}
			return
		}

		// Get hall cost
		var hall models.Hall
		if err := h.Conf.Db.First(&hall, reservation.HallID).Error; err != nil {
			SendError(c, HALL_NOT_FOUND, nil)
			return
		}

		// Compute cost
		reservation.CalculateTotalCost(hall.CostPerDay)

		// Save reservation using generic repo
		if err := h.Repo.Create(&reservation); err != nil {
			SendError(c, FAILED_CREATE_RESERVATION, err)
			return
		}

		// Generate receipt
		if err := receipt.GenerateReceipt(&reservation); err != nil {
			SendError(c, FAILED_CREATE_RECEIPT, err)
			return
		}

		// Calculate and send duration/cost breakdown
		duration := int(reservation.EndDate.Sub(reservation.StartDate).Hours() / 24)
		if duration < 1 {
			duration = 1
		}

		SendSuccessBody(c, gin.H{
			"reservation": reservation,
			"details": gin.H{
				"duration_days": duration,
				"cost_per_day":  reservation.TotalCost / float64(duration),
			},
		})
	}
}

// UpdateReservation modifies an existing reservation
func (h *ReservationHandler) UpdateReservation() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var reservation models.Reservation
		if err := h.Repo.GetByID(id, &reservation); err != nil {
			SendError(c, RESERVATION_NOT_FOUND, err)
			return
		}

		var updated models.Reservation
		if err := c.ShouldBindJSON(&updated); err != nil {
			SendError(c, INVALID_REQ_PAYLOAD, err)
			return
		}

		if !updated.StartDate.Before(updated.EndDate) {
			SendError(c, INVALID_RES_START_DATE, nil)
			return
		}

		var count int64
		h.Conf.Db.Model(&models.Reservation{}).
			Where("hall_id = ? AND id != ? AND ((start_date BETWEEN ? AND ?) OR (end_date BETWEEN ? AND ?))",
				updated.HallID, id,
				updated.StartDate, updated.EndDate,
				updated.StartDate, updated.EndDate).
			Count(&count)

		if count > 0 {
			SendError(c, HALL_BOOKED, nil)
			return
		}

		var hall models.Hall
		if err := h.Conf.Db.First(&hall, updated.HallID).Error; err != nil {
			SendError(c, RESERVATION_NOT_FOUND, err)
			return
		}

		reservation.Name = updated.Name
		reservation.Company = updated.Company
		reservation.HallID = updated.HallID
		reservation.StartDate = updated.StartDate
		reservation.EndDate = updated.EndDate
		reservation.CalculateTotalCost(hall.CostPerDay)

		if err := h.Repo.Update(id, &reservation); err != nil {
			SendError(c, FAILED_UPDATE_RESERVATION, err)
			return
		}

		SendSuccessBody(c, reservation)
	}
}

// GetReservations retrieves reservations with enhanced filtering and sorting.
func (h *ReservationHandler) GetReservations() gin.HandlerFunc {
	return func(c *gin.Context) {
		var reservations []models.Reservation
		query := h.Repo.DB

		if dateStr := c.Query("date"); dateStr != "" {
			if parsedDate, err := time.Parse("2006-01-02", dateStr); err == nil {
				query = query.Where("start_date <= ? AND end_date >= ?", parsedDate, parsedDate)
			}
		}
		if company := c.Query("company"); company != "" {
			query = query.Where("LOWER(company) = ?", strings.ToLower(company))
		}
		if hall := c.Query("hall"); hall != "" {
			query = query.Where("hall_id = ?", hall)
		}
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

		if err := query.Find(&reservations).Error; err != nil {
			SendError(c, FAILED_GET_RESERVATIONS, err)
			return
		}
		if len(reservations) == 0 {
			SendError(c, NO_RESERVATIONS, nil)
			return
		}
		SendSuccessBody(c, reservations)
	}
}

// DeleteReservation removes a reservation and its receipt file
func (h *ReservationHandler) DeleteReservation() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Fetch using repo
		var reservation models.Reservation
		if err := h.Repo.GetByID(id, &reservation); err != nil {
			SendError(c, RESERVATION_NOT_FOUND, err)
			return
		}

		// Delete using repo
		if err := h.Repo.Delete(id); err != nil {
			SendError(c, FAILED_DELETE_RESERVATION, err)
			return
		}

		// Remove the associated receipt file
		if err := deleteReceiptFile(reservation.ID); err != nil {
			SendError(c, FAILED_DELETE_RECEIPT, err)
			return
		}

		SendSuccess(c)
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
func (h *ReservationHandler) GetCategorizedReservations() gin.HandlerFunc {
	return func(c *gin.Context) {
		var reservations []models.Reservation

		if err := h.Repo.DB.Preload("Hall").Find(&reservations).Error; err != nil {
			SendError(c, FAILED_GET_RESERVATIONS, err)
			return
		}

		now := time.Now()
		var past, current, upcoming []models.Reservation

		for _, r := range reservations {
			if r.EndDate.Before(now) {
				past = append(past, r)
			} else if r.StartDate.After(now) {
				upcoming = append(upcoming, r)
			} else {
				current = append(current, r)
			}
		}

		SendSuccessBody(c, gin.H{
			"past":     past,
			"current":  current,
			"upcoming": upcoming,
		})
	}
}

// GetReservationSummary aggregates reservation data for dashboard display.
func (h *ReservationHandler) GetReservationSummary() gin.HandlerFunc {
	return func(c *gin.Context) {
		var reservations []models.Reservation

		if h.Conf.Db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database is disabled."})
			return
		}

		if err := h.Repo.GetAll(&reservations); err != nil {
			SendError(c, FAILED_GET_RESERVATIONS, err)
			return
		}

		now := time.Now()
		var past, current, upcoming int
		var revenue float64

		for _, r := range reservations {
			revenue += r.TotalCost
			switch {
			case r.EndDate.Before(now):
				past++
			case r.StartDate.After(now):
				upcoming++
			default:
				current++
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"total_reservations":    len(reservations),
			"past_reservations":     past,
			"current_reservations":  current,
			"upcoming_reservations": upcoming,
			"total_revenue":         revenue,
		})
	}
}
