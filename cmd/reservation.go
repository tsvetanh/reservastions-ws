package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"storage/configuration"
	"storage/models"
	"time"
)

// Declare the main reservation command
var ReservationCmd = &cobra.Command{
	Use:   "reservation",
	Short: "Manage reservations",
}

// Declare variables for flags
var reservationName string
var reservationCompany string
var reservationHallID int
var reservationStartDate string
var reservationEndDate string

// Create Reservation Command
var createReservationCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new reservation",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := configuration.Init()
		if err != nil {
			fmt.Println("Failed to initialize configuration:", err)
			return
		}

		// Convert date strings to time.Time
		startDate, err := time.Parse("2006-01-02", reservationStartDate)
		if err != nil {
			fmt.Println("Invalid start date format. Expected YYYY-MM-DD")
			return
		}
		endDate, err := time.Parse("2006-01-02", reservationEndDate)
		if err != nil {
			fmt.Println("Invalid end date format. Expected YYYY-MM-DD")
			return
		}

		// Create reservation object
		reservation := models.Reservation{
			Name:      reservationName,
			Company:   reservationCompany,
			HallID:    uint(reservationHallID), // Ensure proper type conversion
			StartDate: startDate,
			EndDate:   endDate,
		}

		// Save reservation in the database
		if err := conf.Db.Create(&reservation).Error; err != nil {
			fmt.Println("Failed to create reservation:", err)
		} else {
			fmt.Println("Reservation created successfully:", reservation)
		}
	},
}

var listReservationsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all reservations",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := configuration.Init()
		if err != nil {
			fmt.Println("Failed to initialize configuration:", err)
			return
		}

		var reservations []models.Reservation
		if err := conf.Db.Find(&reservations).Error; err != nil {
			fmt.Println("Failed to retrieve reservations:", err)
			return
		}

		fmt.Println("\nReservations")
		fmt.Println("-------------------------------------------")
		fmt.Printf("%-5s %-15s %-15s %-5s %-10s %-10s %-10s\n", "ID", "Name", "Company", "Hall", "Start Date", "End Date", "Total Cost")
		fmt.Println("-------------------------------------------")

		for _, r := range reservations {
			fmt.Printf("%-5d %-15s %-15s %-5d %-10s %-10s %-10.2f\n", r.ID, r.Name, r.Company, r.HallID, r.StartDate.Format("2006-01-02"), r.EndDate.Format("2006-01-02"), r.TotalCost)
		}
		fmt.Println("-------------------------------------------")
	},
}

func init() {
	ReservationCmd.AddCommand(listReservationsCmd)
}

// Initialize command and flags
func init() {
	createReservationCmd.Flags().StringVarP(&reservationName, "name", "n", "", "Reservation Name")
	createReservationCmd.Flags().StringVarP(&reservationCompany, "company", "c", "", "Company Name")
	createReservationCmd.Flags().IntVar(&reservationHallID, "hall", 0, "Hall ID")
	createReservationCmd.Flags().StringVarP(&reservationStartDate, "start", "s", "", "Start Date (YYYY-MM-DD)")
	createReservationCmd.Flags().StringVarP(&reservationEndDate, "end", "e", "", "End Date (YYYY-MM-DD)")

	// Mark required flags
	createReservationCmd.MarkFlagRequired("name")
	createReservationCmd.MarkFlagRequired("company")
	createReservationCmd.MarkFlagRequired("hall")
	createReservationCmd.MarkFlagRequired("start")
	createReservationCmd.MarkFlagRequired("end")

	// Attach commands to ReservationCmd
	ReservationCmd.AddCommand(createReservationCmd)
}
