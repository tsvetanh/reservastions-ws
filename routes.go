package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"storage/configuration"
	. "storage/internal/handlers"
	"storage/internal/services"
	. "storage/middleware"
)

func Routes(d *configuration.Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(LoggingMiddleware)
	r.Use(CORSandCSP())

	r.GET("/version", func(c *gin.Context) {
		c.String(http.StatusOK, "This is version 3.0 - updates: Full code refactoring; Standardize response structure and error handling")
	})

	apiGroup := r.Group("/api")
	{
		// Public routes
		apiGroup.POST("/login", LoginHandler(d))
		// Register route
		apiGroup.POST("/register", RegisterHandler(d))

		// Routes requiring authentication
		protected := apiGroup.Group("/")
		protected.GET("/halls/image/:name", ServeImage())

		protected.Use(AuthMiddleware(d))

		{ // Users Routes
			usersGroup := protected.Group("/")
			usersGroup.Use(AllowedRoles("admin"))

			usersGroup.GET("/users", HandlerGetAllUsers(d))
			usersGroup.GET("/roles", HandlerGetAllRoles(d))
			usersGroup.POST("/add-role", HandlerInsertRole(d))
			usersGroup.POST("/update-role", HandlerUpdateRole(d))
			usersGroup.POST("/assign-role", HandlerAssignRole(d))
			usersGroup.POST("/revoke-role", HandlerRevokeRole(d))
		}

		{ // Hall Management Routes
			hallGroup := protected.Group("/halls")
			hallGroup.Use(AllowedRoles("user"))

			hallGroup.POST("", CreateHall(d))                                     // Create a new hall
			hallGroup.GET("", GetHalls(d))                                        // Get all halls
			hallGroup.PUT("/:id", UpdateHall(d))                                  // Update a hall by ID
			hallGroup.DELETE("/:id", DeleteHall(d))                               // Delete a hall by ID
			hallGroup.GET("/:id/utilization", services.GetHallUtilizationRate(d)) // Statistics on Hall usage
		}

		{ // Reservation Management Routes
			reservationGroup := protected.Group("/reservations")
			reservationGroup.Use(AllowedRoles("user"))

			reservationGroup.POST("", CreateReservation(d))                     // Create a new reservation
			reservationGroup.GET("", GetReservations(d))                        // Get all reservations
			reservationGroup.DELETE("/:id", DeleteReservation(d))               // Delete a reservation by ID
			reservationGroup.PUT("/:id", UpdateReservation(d))                  // Manage/Modify reservations
			reservationGroup.GET("/categorized", GetCategorizedReservations(d)) // New endpoint for categorized reservations.
			reservationGroup.GET("/summary", GetReservationSummary(d))          // Dashboard for reservations
		}
	}

	return r

}
