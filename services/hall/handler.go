package hall

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"storage/configuration"
	"storage/models"
)

// CreateHall handles the creation of a new hall.
func CreateHall(conf *configuration.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		var hall models.Hall
		if err := c.ShouldBindJSON(&hall); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		if hall.Capacity <= 0 || hall.CostPerDay <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Capacity and cost must be positive numbers"})
			return
		}

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
