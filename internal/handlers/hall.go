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
	. "storage/internal/utils"
	"time"
)

// CreateHall handles the creation of a new hall.
func CreateHall(conf *configuration.Dependencies) gin.HandlerFunc {
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
		err = json.Unmarshal([]byte(hallData[0]), &hall)
		if err != nil {
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

		if err := conf.Db.Create(&hall).Error; err != nil {
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

			if err := conf.Db.Create(&image).Error; err != nil {
				SendError(c, FAILED_SAVE_IMAGE_DATA, err)
				return
			}
		}

		SendSuccessBody(c, hall)
	}
}

func ServeImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Param("name")

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

// GetHalls retrieves all available halls.
func GetHalls(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		var halls []models.Hall

		if err := conf.Db.Preload("Reservations").Preload("HallImages").Find(&halls).Error; err != nil {
			SendError(c, FAILED_GET_HALLS, err)
			return
		}

		for i := range halls {
			var imagePaths []string
			for _, image := range halls[i].HallImages {
				imagePaths = append(imagePaths, image.ImageName)
			}
			halls[i].ImageURLs = imagePaths
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

// UpdateHall modifies an existing hall.
func UpdateHall(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var hall models.Hall
		if err := conf.Db.First(&hall, id).Error; err != nil {
			SendError(c, HALL_NOT_FOUND, err)
			return
		}

		if err := c.ShouldBindJSON(&hall); err != nil {
			SendError(c, INVALID_REQ_PAYLOAD, err)
			return
		}

		conf.Db.Save(&hall)
		SendSuccessBody(c, hall)
	}
}

// DeleteHall removes a hall.
func DeleteHall(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := conf.Db.Delete(&models.Hall{}, id).Error; err != nil {
			SendError(c, FAILED_DELETE_HALL, err)
			return
		}

		SendSuccess(c)
	}
}
