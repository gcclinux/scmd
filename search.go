package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// search prints the result returned from PostgreSQL database
func search(pattern string) {

	WriteLogToFile(webLog, "CLI: "+pattern)

	// Initialize database connection
	if err := InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer CloseDB()

	// Search commands
	received, err := SearchCommands(pattern, "json")
	if err != nil {
		log.Fatalf("Error searching commands: %v", err)
	}

	var dt []CommandRecord
	json.Unmarshal(received, &dt)

	checkDB(received)

	for x := range dt {
		cmd := string(dt[x].Key)
		check := isCode(dt[x].Key)
		if check {
			if !strings.HasSuffix(cmd, "{{end}}") {
				cmd = replaceLast(cmd, "}", "\n}")
			}
			cmd = strings.ReplaceAll(cmd, "\n\t\n\t", "\n\t\t")
			fmt.Printf("\nID:%v - %v", dt[x].Id, dt[x].Data)
			fmt.Println()
			fmt.Println(cmd)
			fmt.Println()
		} else {
			// Pretty print the result
			out, _ := json.MarshalIndent(&dt[x], "", "  ")
			fmt.Println(string(out))
		}
	}
}
