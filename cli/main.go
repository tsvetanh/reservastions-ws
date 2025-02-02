package main

import (
	"fmt"
	"os"
	"storage/cmd" // Import CLI commands
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
