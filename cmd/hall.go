package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"storage/configuration"
	"storage/models"
	"time"
)

// Hall command group
var HallCmd = &cobra.Command{
	Use:   "hall",
	Short: "Manage halls",
}

// Create Hall
var createHallCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new hall",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := configuration.Init()
		if err != nil {
			fmt.Println("Failed to initialize configuration:", err)
			return
		}

		hall := models.Hall{
			ID:            uint(hallID), //
			Capacity:      hallCapacity,
			CostPerDay:    hallCost,
			AvailableFrom: parseTime(hallAvailableFrom),
			AvailableTo:   parseTime(hallAvailableTo),
		}

		// Check if hall already exists (to prevent duplicate primary key errors)
		var existingHall models.Hall
		if err := conf.Db.First(&existingHall, hall.ID).Error; err == nil {
			fmt.Println("Error: A hall with this ID already exists.")
			return
		}

		if err := conf.Db.Create(&hall).Error; err != nil {
			fmt.Println("Failed to create hall:", err)
		} else {
			fmt.Println("Hall created successfully:", hall)
		}
	},
}

var listHallsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all halls",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := configuration.Init()
		if err != nil {
			fmt.Println("Failed to initialize configuration:", err)
			return
		}

		var halls []models.Hall
		if err := conf.Db.Find(&halls).Error; err != nil {
			fmt.Println("Failed to retrieve halls:", err)
			return
		}

		fmt.Println("\n🏢 Available Halls")
		fmt.Println("-------------------------------------------")
		fmt.Printf("%-5s %-10s %-10s\n", "ID", "Capacity", "Cost/Day")
		fmt.Println("-------------------------------------------")

		for _, h := range halls {
			fmt.Printf("%-5d %-10d $%-9.2f\n", h.ID, h.Capacity, h.CostPerDay)
		}
		fmt.Println("-------------------------------------------")
	},
}

var deleteHallCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a hall",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := configuration.Init()
		if err != nil {
			fmt.Println("Failed to initialize configuration:", err)
			return
		}

		if err := conf.Db.Delete(&models.Hall{}, hallID).Error; err != nil {
			fmt.Println("Failed to delete hall:", err)
		} else {
			fmt.Println("Hall deleted successfully")
		}
	},
}

// CLI flags
var hallID int
var hallCapacity int
var hallCost float64
var hallAvailableFrom, hallAvailableTo string

func init() {
	HallCmd.AddCommand(createHallCmd)
	HallCmd.AddCommand(listHallsCmd)
	HallCmd.AddCommand(deleteHallCmd)

	// Add flag to manually set the hall ID
	createHallCmd.Flags().IntVarP(&hallID, "id", "i", 0, "Hall ID (Optional)")
	createHallCmd.Flags().IntVarP(&hallCapacity, "capacity", "c", 0, "Hall Capacity")
	createHallCmd.Flags().Float64VarP(&hallCost, "cost", "p", 0, "Cost Per Day")
	createHallCmd.Flags().StringVarP(&hallAvailableFrom, "from", "f", "", "Available From (YYYY-MM-DD)")
	createHallCmd.Flags().StringVarP(&hallAvailableTo, "to", "t", "", "Available To (YYYY-MM-DD)")
	createHallCmd.MarkFlagRequired("capacity")
	createHallCmd.MarkFlagRequired("cost")

	deleteHallCmd.Flags().IntVarP(&hallID, "id", "d", 0, "Hall ID")
	deleteHallCmd.MarkFlagRequired("id")
}

// parseTime converts a string (YYYY-MM-DD) into a time.Time object
func parseTime(dateStr string) time.Time {
	parsedTime, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		fmt.Println("Invalid date format. Expected YYYY-MM-DD")
		return time.Time{}
	}
	return parsedTime
}
