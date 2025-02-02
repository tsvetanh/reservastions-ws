package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"storage/configuration"
	. "storage/middleware"
	"storage/services/hall" // Import Hall service
	login "storage/services/login"
	register "storage/services/register"
	"storage/services/reservation" // Import Reservation service
	"storage/services/user"
)

func Routes(d *configuration.Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(LoggingMiddleware)
	r.Use(CORSandCSP())

	r.GET("/version", func(c *gin.Context) {
		c.String(http.StatusOK, "This is version 2.0 - updates:LOGIN authentication JWT added.")
	})

	apiGroup := r.Group("/api")
	{
		// Public routes
		apiGroup.POST("/login", login.LoginHandler(d))
		// Register route
		apiGroup.POST("/register", register.RegisterHandler(d))

		// Routes requiring authentication
		protected := apiGroup.Group("/")
		protected.Use(AuthMiddleware(d))
		protected.Use(AllowedRoles("admin"))

		{ // Users Routes
			usersGroup := protected.Group("/")

			usersGroup.GET("/users", user.HandlerGetAllUsers(d))
			usersGroup.GET("/roles", user.HandlerGetAllRoles(d))
			usersGroup.POST("/add-role", user.HandlerInsertRole(d))
			usersGroup.POST("/update-role", user.HandlerUpdateRole(d))
			usersGroup.POST("/assign-role", user.HandlerAssignRole(d))
			usersGroup.POST("/revoke-role", user.HandlerRevokeRole(d))
		}

		{ // Hall Management Routes
			hallGroup := protected.Group("/halls")
			hallGroup.Use(AllowedRoles("user"))

			hallGroup.POST("", hall.CreateHall(d))                            // Create a new hall
			hallGroup.GET("", hall.GetHalls(d))                               // Get all halls
			hallGroup.GET("/image/:path", hall.ServeImage())                  // Get all halls
			hallGroup.PUT("/:id", hall.UpdateHall(d))                         // Update a hall by ID
			hallGroup.DELETE("/:id", hall.DeleteHall(d))                      // Delete a hall by ID
			hallGroup.GET("/:id/utilization", hall.GetHallUtilizationRate(d)) // Statistics on Hall usage
		}

		{ // Reservation Management Routes
			reservationGroup := protected.Group("/reservations")
			reservationGroup.Use(AllowedRoles("user"))

			reservationGroup.POST("", reservation.CreateReservation(d))                     // Create a new reservation
			reservationGroup.GET("", reservation.GetReservations(d))                        // Get all reservations
			reservationGroup.DELETE("/:id", reservation.DeleteReservation(d))               // Delete a reservation by ID
			reservationGroup.PUT("/:id", reservation.UpdateReservation(d))                  //Manage/Modify reservations
			reservationGroup.GET("/categorized", reservation.GetCategorizedReservations(d)) // New endpoint for categorized reservations.
			reservationGroup.GET("/summary", reservation.GetReservationSummary(d))          //Dashboard for reservations
		}
	}

	return r

}
