package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"storage/configuration"
	"storage/internal/models"
	"storage/internal/repos"
	. "storage/internal/utils"
	"time"
)

type HallHandler struct {
	Repo *repos.BaseRepository[models.Hall]
	Conf *configuration.Dependencies
}

func NewHallHandler(conf *configuration.Dependencies) *HallHandler {
	return &HallHandler{
		Repo: &repos.BaseRepository[models.Hall]{DB: conf.Db},
		Conf: conf,
	}
}

// CreateHall handles the creation of a new hall with optional image upload.
func (h *HallHandler) CreateHall() gin.HandlerFunc {
	return func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			SendError(c, INVALID_REQ_PAYLOAD, err)
			return
		}

		hallData := form.Value["hall"]
		if len(hallData) == 0 {
			SendError(c, MISSING_HALL_DATA, nil)
			return
		}

		var hall models.Hall
		if err := json.Unmarshal([]byte(hallData[0]), &hall); err != nil {
			SendError(c, INVALID_HALL_DATA, err)
			return
		}

		if hall.Capacity <= 0 || hall.CostPerDay <= 0 {
			SendError(c, CAPACITY_AND_COST_ERROR, nil)
			return
		}

		if !hall.AvailableFrom.IsZero() && !hall.AvailableTo.IsZero() {
			if !hall.AvailableFrom.Before(hall.AvailableTo) {
				SendError(c, INVALID_DATE_RANGE, nil)
				return
			}
			if hall.AvailableFrom.Before(time.Now()) {
				SendError(c, AVAILABLE_FROM_IN_PAST, nil)
				return
			}
		}

		if err := h.Repo.Create(&hall); err != nil {
			SendError(c, FAILED_CREATE_HALL, err)
			return
		}

		files := form.File["images"]
		for _, file := range files {
			filename := fmt.Sprintf("%d_%s", hall.ID, file.Filename)
			if err := c.SaveUploadedFile(file, "uploads/"+filename); err != nil {
				SendError(c, FAILED_SAVE_IMAGE, err)
				return
			}

			image := models.HallImage{
				HallID:    hall.ID,
				ImageName: filename,
			}
			if err := h.Conf.Db.Create(&image).Error; err != nil {
				SendError(c, FAILED_SAVE_IMAGE_DATA, err)
				return
			}
		}

		SendSuccessBody(c, hall)
	}
}

// AddHallImages handles the uploading new images for a hall
func (h *HallHandler) AddHallImages() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var hall models.Hall
		if err := h.Repo.GetByID(id, &hall); err != nil {
			SendError(c, HALL_NOT_FOUND, err)
			return
		}

		form, err := c.MultipartForm()
		if err != nil {
			SendError(c, INVALID_REQ_PAYLOAD, err)
			return
		}

		files := form.File["images"]
		if len(files) == 0 {
			SendError(c, MISSING_IMAGES, nil)
			return
		}

		var savedImages []models.HallImage
		for _, file := range files {
			filename := fmt.Sprintf("%d_%s", hall.ID, file.Filename)

			savePath := filepath.Join("uploads", filename)
			if err := c.SaveUploadedFile(file, savePath); err != nil {
				SendError(c, FAILED_SAVE_IMAGE, err)
				return
			}

			image := models.HallImage{
				HallID:    hall.ID,
				ImageName: filename,
			}

			if err := h.Conf.Db.Create(&image).Error; err != nil {
				SendError(c, FAILED_SAVE_IMAGE_DATA, err)
				return
			}

			savedImages = append(savedImages, image)
		}

		SendSuccessBody(c, gin.H{
			"message": "Images uploaded successfully",
			"count":   len(savedImages),
			"images":  savedImages,
		})
	}
}

func CreateHall_old(conf *configuration.Dependencies) gin.HandlerFunc {
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

// GetHall retrieves hall with reservations and images.
func (h *HallHandler) GetHall() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var hall models.Hall
		if err := h.Repo.GetByID(id, &hall, "Reservations", "HallImages"); err != nil {
			SendError(c, HALL_NOT_FOUND, err)
			return
		}

		// Populate image URLs
		for _, image := range hall.HallImages {
			hall.ImageURLs = append(hall.ImageURLs, image.ImageName)
		}

		SendSuccessBody(c, hall)
	}
}

// GetHalls retrieves all halls with reservations and images.
func (h *HallHandler) GetHalls() gin.HandlerFunc {
	return func(c *gin.Context) {
		var halls []models.Hall
		err := h.Repo.GetAll(&halls, "Reservations", "HallImages")
		if err != nil {
			SendError(c, FAILED_GET_HALLS, err)
			return
		}

		for i := range halls {
			for _, image := range halls[i].HallImages {
				halls[i].ImageURLs = append(halls[i].ImageURLs, image.ImageName)
			}
		}

		SendSuccessBody(c, halls)
	}
}

func GetHalls_old(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		var halls []models.Hall
		if err := conf.Db.Find(&halls).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve halls"})
			return
		}
		c.JSON(http.StatusOK, halls)
	}
}

// UpdateHall updates a hall by ID.
func (h *HallHandler) UpdateHall() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var existing models.Hall
		if err := h.Repo.GetByID(id, &existing); err != nil {
			SendError(c, HALL_NOT_FOUND, err)
			return
		}

		if err := c.ShouldBindJSON(&existing); err != nil {
			SendError(c, INVALID_REQ_PAYLOAD, err)
			return
		}

		if err := h.Repo.Update(id, &existing); err != nil {
			SendError(c, FAILED_UPDATE_HALL, err)
			return
		}

		SendSuccessBody(c, existing)
	}
}

// DeleteHall deletes a hall by ID.
func (h *HallHandler) DeleteHall() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := h.Repo.Delete(id); err != nil {
			SendError(c, FAILED_DELETE_HALL, err)
			return
		}

		SendSuccess(c)
	}
}

// ServeImage retrieves an image
func ServeImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Param("name")

		if path == "default.png" {
			SendSuccess(c)
			return
		}

		imagePath := filepath.Join("uploads/", path)

		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			SendError(c, IMAGE_NOT_FOUND, err)
			return
		}

		ext := filepath.Ext(path)
		var contentType string
		switch ext {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		default:
			contentType = "application/octet-stream"
		}

		c.Header("Content-Type", contentType)

		c.File(imagePath)
	}
}
