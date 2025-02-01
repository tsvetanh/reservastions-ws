package reservation

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// GenerateReceipt creates a .txt file with reservation details
func GenerateReceipt(reservation *Reservation) error {
	// Ensure the "receipt" directory exists
	receiptDir := "receipt"
	if _, err := os.Stat(receiptDir); os.IsNotExist(err) {
		err := os.Mkdir(receiptDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create receipt directory: %v", err)
		}
	}

	// Generate filename using reservation ID
	filename := fmt.Sprintf("receipt_%d.txt", reservation.ID)
	filePath := filepath.Join(receiptDir, filename)

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create receipt file: %v", err)
	}
	defer file.Close()

	// Write reservation details to the file
	content := fmt.Sprintf(
		"Reservation Receipt\n"+
			"--------------------\n"+
			"Reservation ID: %d\n"+
			"Name: %s\n"+
			"Company: %s\n"+
			"Hall ID: %d\n"+
			"Start Date: %s\n"+
			"End Date: %s\n"+
			"Total Cost: %.2f BGN\n"+
			"--------------------\n"+
			"Generated on: %s\n",
		reservation.ID,
		reservation.Name,
		reservation.Company,
		reservation.HallID,
		reservation.StartDate.Format("2006-01-02"),
		reservation.EndDate.Format("2006-01-02"),
		reservation.TotalCost,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to receipt file: %v", err)
	}

	fmt.Println("Receipt generated:", filePath)
	return nil
}

// ensureDirectoryExists checks if a directory exists and creates it if not
func ensureDirectoryExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, 0755)
		if err != nil {
			return fmt.Errorf("error creating directory %s: %v", dir, err)
		}
	}
	return nil
}
