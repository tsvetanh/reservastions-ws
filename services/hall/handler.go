package hall

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"reservations/configuration"
	"time"
)

// CreateHall handles the creation of a new hall.

func CreateHall(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Step 1: Get the JSON part of the request (hall data)
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
			return
		}

		// The hall data will be in the 'hall' field of the multipart form
		hallData := form.Value["hall"]
		if len(hallData) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Hall data missing"})
			return
		}

		// Step 2: Unmarshal the hall data from the form field into the Hall struct
		var hall Hall
		err = json.Unmarshal([]byte(hallData[0]), &hall)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hall JSON data"})
			return
		}

		// Step 3: Validate the hall data
		if hall.Capacity <= 0 || hall.CostPerDay <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Capacity and cost must be positive numbers"})
			return
		}

		if !hall.AvailableFrom.IsZero() && !hall.AvailableTo.IsZero() {
			if !hall.AvailableFrom.Before(hall.AvailableTo) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "AvailableFrom must be before AvailableTo"})
				return
			}
			if hall.AvailableFrom.Before(time.Now()) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "AvailableFrom cannot be in the past"})
				return
			}
		}

		// Step 4: Create the hall in the database
		if err := conf.Db.Create(&hall).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create hall"})
			return
		}

		// Step 5: Handle the image uploads (multipart form data)
		files := form.File["images"] // "images" is the form field name for file uploads
		for _, file := range files {
			// Save the image with a unique filename
			filename := fmt.Sprintf("%d_%s", hall.ID, file.Filename)
			filePath := filepath.Join("uploads", filename)

			// Save the file to disk
			if err := c.SaveUploadedFile(file, filePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
				return
			}

			// Step 6: Save the image information to the database
			image := HallImage{
				HallID:   hall.ID,
				ImageURL: filePath, // Or store a URL if you're using cloud storage
			}

			if err := conf.Db.Create(&image).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image data"})
				return
			}
		}

		// Step 7: Return the created hall
		c.JSON(http.StatusOK, hall)
	}
}

func CreateHall_old(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		var hall Hall
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
		var halls []Hall

		// Step 1: Preload the images for each hall.
		if err := conf.Db.Preload("Reservations").Preload("HallImages").Find(&halls).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve halls"})
			return
		}

		// Step 2: Add image URLs as an array of strings for each hall
		for i := range halls {
			var imagePaths []string
			for _, image := range halls[i].HallImages {
				imagePaths = append(imagePaths, image.ImageURL) // Add the image URL to the array
			}
			// Add the image URLs to the hall object
			halls[i].ImageURLs = imagePaths
		}

		// Step 3: Return the halls with the image URLs
		c.JSON(http.StatusOK, halls)
	}
}

func GetHalls_old(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		var halls []Hall
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

		var hall Hall
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

		if err := conf.Db.Delete(&Hall{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete hall"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Hall deleted successfully"})
	}
}
