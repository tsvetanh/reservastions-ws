package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"storage/configuration"
	"storage/models"
	"time"
)

var SummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show reservation summary",
	Run: func(cmd *cobra.Command, args []string) {
		conf, _ := configuration.Init()

		// Retrieve reservations
		var reservations []models.Reservation
		if err := conf.Db.Find(&reservations).Error; err != nil {
			fmt.Println("Failed to retrieve reservations:", err)
			return
		}

		now := time.Now()
		var pastCount, currentCount, upcomingCount int
		var totalRevenue float64

		for _, r := range reservations {
			totalRevenue += r.TotalCost
			if r.EndDate.Before(now) {
				pastCount++
			} else if r.StartDate.After(now) {
				upcomingCount++
			} else {
				currentCount++
			}
		}

		fmt.Println("\n Reservation Summary")
		fmt.Println("---------------------------------------------")
		fmt.Printf("Total Reservations:    %d\n", len(reservations))
		fmt.Printf("Past Reservations:     %d\n", pastCount)
		fmt.Printf("Current Reservations:  %d\n", currentCount)
		fmt.Printf("Upcoming Reservations: %d\n", upcomingCount)
		fmt.Println("---------------------------------------------")
		fmt.Printf("Total Revenue:         $%.2f\n", totalRevenue)
		fmt.Println("---------------------------------------------")
	},
}
