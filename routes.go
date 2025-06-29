package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"storage/configuration"
	"storage/internal/handlers"
	"storage/internal/services"
	"storage/middleware"
)

func Routes(d *configuration.Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.CORSandCSP())

	r.GET("/version", func(c *gin.Context) {
		c.String(http.StatusOK, "This is version 3.0 - updates: Full code refactoring; Standardize response structure and error handling")
	})

	apiGroup := r.Group("/api")
	{
		// Public routes
		apiGroup.POST("/login", handlers.LoginHandler(d))
		apiGroup.POST("/register", handlers.RegisterHandler(d))

		apiGroup.GET("/halls/images/:name", handlers.ServeImage()) // Get image by name

		// Authenticated routes
		userGroup := apiGroup.Group("/")
		userGroup.Use(middleware.AuthMiddleware(d))
		userGroup.Use(middleware.AllowedRoles("user"))

		registerHallRoutes(userGroup, d)
		registerReservationRoutes(userGroup, d)

		// Admin-only routes
		adminGroup := userGroup.Group("/admin")
		adminGroup.Use(middleware.AllowedRoles("admin"))

		registerUserRoleRoutes(adminGroup, d)

	}

	return r
}

// --- Reservation Routes ---
func registerReservationRoutes(r *gin.RouterGroup, d *configuration.Dependencies) {
	resHandler := handlers.NewReservationHandler(d)
	resGroup := r.Group("/reservations")

	resGroup.POST("", resHandler.CreateReservation())                     // Create a new reservation
	resGroup.GET("", resHandler.GetReservations())                        // Get all reservations
	resGroup.DELETE("/:id", resHandler.DeleteReservation())               // Delete a reservation by ID
	resGroup.PUT("/:id", resHandler.UpdateReservation())                  // Manage/Modify reservations
	resGroup.GET("/categorized", resHandler.GetCategorizedReservations()) // New endpoint for categorized reservations.
	resGroup.GET("/summary", resHandler.GetReservationSummary())          // Dashboard for reservations

}

// --- User & Role Routes ---
func registerUserRoleRoutes(r *gin.RouterGroup, d *configuration.Dependencies) {
	userHandler := handlers.NewUserHandler(d)

	// Users
	r.GET("/users", userHandler.GetAllUsers()) // Get all users

	// Roles
	r.GET("/roles", userHandler.GetAllRoles())    // Get all roles
	r.POST("/roles", userHandler.InsertRole())    // Create new role
	r.PUT("/roles/:id", userHandler.UpdateRole()) // Update role

	// Role assignment
	r.POST("/users/:id/roles", userHandler.AssignRole())         // Assign role to a user
	r.DELETE("/users/:id/roles/:role", userHandler.RevokeRole()) // Revoke role from a user
}

// --- Hall Routes ---
func registerHallRoutes(r *gin.RouterGroup, d *configuration.Dependencies) {
	hallHandler := handlers.NewHallHandler(d)
	hallGroup := r.Group("/halls")

	hallGroup.POST("", hallHandler.CreateHall())       // Create a new hall
	hallGroup.GET("", hallHandler.GetHalls())          // Get all halls
	hallGroup.GET("/:id", hallHandler.GetHall())       // Get a hall by ID
	hallGroup.PUT("/:id", hallHandler.UpdateHall())    // Update a hall by ID
	hallGroup.DELETE("/:id", hallHandler.DeleteHall()) // Delete a hall by ID

	hallGroup.POST("/:id/images", hallHandler.AddHallImages())            // Upload images for a hall
	hallGroup.GET("/:id/utilization", services.GetHallUtilizationRate(d)) // Statistics on Hall usage
}
