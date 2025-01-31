package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"storage/configuration"
	"storage/middleware"
	"storage/services/hall" // Import Hall service
	login "storage/services/login"
	register "storage/services/register"
	"storage/services/reservation" // Import Reservation service
	"storage/services/user"
)

func Routes(d *configuration.Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.CORSandCSP())

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
		//protected.Use(middleware.AuthMiddleware())

		// Users route
		protected.GET("/users", user.HandlerGetAllUsers(d))
		protected.POST("/add-role", user.HandlerInsertRole(d))
		protected.POST("/update-role", user.HandlerUpdateRole(d))
		protected.POST("/assign-role", user.HandlerAssignRole(d))
		protected.POST("/revoke-role", user.HandlerRevokeRole(d))

		// Hall Management Routes
		protected.POST("/halls", hall.CreateHall(d))       // Create a new hall
		protected.GET("/halls", hall.GetHalls(d))          // Get all halls
		protected.PUT("/halls/:id", hall.UpdateHall(d))    // Update a hall by ID
		protected.DELETE("/halls/:id", hall.DeleteHall(d)) // Delete a hall by ID

		// Reservation Management Routes
		protected.POST("/reservations", reservation.CreateReservation(d))       // Create a new reservation
		protected.GET("/reservations", reservation.GetReservations(d))          // Get all reservations
		protected.DELETE("/reservations/:id", reservation.DeleteReservation(d)) // Delete a reservation by ID
		protected.PUT("/reservations/:id", reservation.UpdateReservation(d))    //Manage/Modify reservations
	}

	return r

}
