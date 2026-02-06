package main

import (
	"fmt"
	"log"
)

// saveCMD takes the cmd and details (string) and saves it to PostgreSQL database
func saveCmd(cmd, details string, db string) {
	// Initialize database connection
	if err := InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer CloseDB()

	// Add command to database
	status, err := AddCommand(cmd, details)
	if err != nil {
		fmt.Println("Error saving command:", err)
		fmt.Println("returned: ( false )")
	} else {
		fmt.Println("returned: (", status, ")")
	}
	fmt.Println()
}
