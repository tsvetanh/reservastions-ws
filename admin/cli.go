package admin

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"storage/configuration"
	"storage/models"
)

// InitConfig initializes the configuration only when needed.
func InitConfig() (*configuration.Dependencies, error) {
	conf, err := configuration.Init()
	if err != nil {
		return nil, fmt.Errorf("error initializing configuration: %w", err)
	}
	return conf, nil
}

// RootCmd is the main command for the CLI.
var RootCmd = &cobra.Command{
	Use:   "admin-cli",
	Short: "Admin CLI for Hall Reservation System",
	Long:  "A command-line interface to manage halls and reservations.",
}

// --- Hall Commands ---

var hallCmd = &cobra.Command{
	Use:   "hall",
	Short: "Manage halls",
	Long:  "Create, list, update, delete halls, and check utilization.",
}

// hallCreateCmd creates a new hall.
var hallCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new hall",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := InitConfig()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		capacity, _ := cmd.Flags().GetInt("capacity")
		location, _ := cmd.Flags().GetString("location")
		available, _ := cmd.Flags().GetBool("available")
		cost, _ := cmd.Flags().GetFloat64("cost")
		availableFromStr, _ := cmd.Flags().GetString("available_from")
		availableToStr, _ := cmd.Flags().GetString("available_to")

		var fromTime, toTime time.Time
		if availableFromStr != "" {
			fromTime, err = time.Parse("2006-01-02", availableFromStr)
			if err != nil {
				fmt.Println("Invalid available_from date:", err)
				return
			}
		}
		if availableToStr != "" {
			toTime, err = time.Parse("2006-01-02", availableToStr)
			if err != nil {
				fmt.Println("Invalid available_to date:", err)
				return
			}
		}

		newHall := models.Hall{
			Capacity:      capacity,
			Location:      location,
			Available:     available,
			CostPerDay:    cost,
			AvailableFrom: fromTime,
			AvailableTo:   toTime,
		}

		if err := conf.Db.Create(&newHall).Error; err != nil {
			fmt.Println("Error creating hall:", err)
			return
		}
		fmt.Println("Hall created successfully:", newHall)
	},
}

// hallListCmd lists all halls.
var hallListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all halls",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := InitConfig()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var halls []models.Hall
		if err := conf.Db.Find(&halls).Error; err != nil {
			fmt.Println("Error fetching halls:", err)
			return
		}
		for _, h := range halls {
			fmt.Printf("ID: %d, Location: %s, Capacity: %d, Cost/Day: %.2f\n",
				h.ID, h.Location, h.Capacity, h.CostPerDay)
		}
	},
}

// hallDeleteCmd deletes a hall.
var hallDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a hall by ID",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Usage: admin-cli hall delete [hall_id]")
			return
		}

		hallID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Invalid hall ID:", err)
			return
		}

		conf, err := InitConfig()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := conf.Db.Delete(&models.Hall{}, hallID).Error; err != nil {
			fmt.Println("Error deleting hall:", err)
			return
		}
		fmt.Println("Hall deleted successfully")
	},
}

// --- Reservation Commands ---

var reservationCmd = &cobra.Command{
	Use:   "reservation",
	Short: "Manage reservations",
	Long:  "Create, list, update, delete reservations, and view summaries.",
}

// reservationCreateCmd creates a new reservation.
var reservationCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new reservation",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := InitConfig()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		name, _ := cmd.Flags().GetString("name")
		company, _ := cmd.Flags().GetString("company")
		startDateStr, _ := cmd.Flags().GetString("start_date")
		endDateStr, _ := cmd.Flags().GetString("end_date")
		hallID, _ := cmd.Flags().GetInt("hall_id")

		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			fmt.Println("Invalid start_date. Use RFC3339 format, e.g., 2025-03-10T09:00:00Z")
			return
		}
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			fmt.Println("Invalid end_date. Use RFC3339 format, e.g., 2025-03-12T18:00:00Z")
			return
		}

		newRes := models.Reservation{
			Name:      name,
			Company:   company,
			StartDate: startDate,
			EndDate:   endDate,
			HallID:    uint(hallID),
		}

		if err := conf.Db.Create(&newRes).Error; err != nil {
			fmt.Println("Error creating reservation:", err)
			return
		}
		fmt.Println("Reservation created successfully:", newRes)
	},
}

// --- CLI Initialization ---

func init() {
	// Add hall commands
	hallCmd.AddCommand(hallCreateCmd)
	hallCmd.AddCommand(hallListCmd)
	hallCmd.AddCommand(hallDeleteCmd)
	RootCmd.AddCommand(hallCmd)

	// Add reservation commands
	reservationCmd.AddCommand(reservationCreateCmd)
	RootCmd.AddCommand(reservationCmd)
}

// Execute runs the CLI application.
func Execute() error {
	return RootCmd.Execute()
}
