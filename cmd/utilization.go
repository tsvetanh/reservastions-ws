package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"storage/configuration"
	"storage/models"
	"time"
)

// Utilization command
var UtilizationCmd = &cobra.Command{
	Use:   "utilization",
	Short: "Show hall utilization rate",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := configuration.Init()
		if err != nil {
			fmt.Println("Failed to initialize configuration:", err)
			return
		}

		// Ensure hall ID is provided
		if hallID == 0 {
			fmt.Println("Error: You must provide a hall ID using --hall")
			return
		}

		// Fetch hall
		var hall models.Hall
		if err := conf.Db.First(&hall, hallID).Error; err != nil {
			fmt.Println("Error: Hall not found")
			return
		}

		// Calculate utilization
		var reservations []models.Reservation
		startDate := time.Now().AddDate(0, 0, -30)
		endDate := time.Now()

		conf.Db.Where("hall_id = ? AND start_date <= ? AND end_date >= ?", hallID, endDate, startDate).Find(&reservations)

		bookedDays := 0
		for _, r := range reservations {
			overlapStart := maxTime(startDate, r.StartDate)
			overlapEnd := minTime(endDate, r.EndDate)
			bookedDays += int(overlapEnd.Sub(overlapStart).Hours() / 24)
		}

		totalDays := int(endDate.Sub(startDate).Hours() / 24)
		utilizationRate := (float64(bookedDays) / float64(totalDays)) * 100

		fmt.Printf("Utilization for Hall %d (Last 30 Days): %.2f%%\n", hallID, utilizationRate)
	},
}

// Utility functions for time comparison
func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// Flags
var HallID int

func init() {
	rootCmd.AddCommand(UtilizationCmd)
	UtilizationCmd.Flags().IntVar(&hallID, "hall", 0, "Hall ID (Required)")
	UtilizationCmd.MarkFlagRequired("hall") // Removed shorthand "-h"
}
