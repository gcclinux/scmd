package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const webLog = "scmdweb.log"

// checkDB will see if the Database exist and if it contains any data.
func checkDB(received []byte) (response string) {

	if (strings.Contains(string(received), "Database") && strings.Contains(string(received), "missing")) || (strings.Contains(string(received), "Database") && strings.Contains(string(received), "empty")) {
		fmt.Println(string(received))
		response = string(received)
	}
	return response
}

// isInt cheking function to see if a string parsed is an integer
func isInt(in string) (result bool) {
	result = true
	_, err := strconv.ParseInt(in, 10, 0)
	if err != nil {
		return false
	}
	return result
}

// copyDB exports all commands from PostgreSQL to a JSON file in the home directory
func copyDB(filename string) {
	fmt.Println()
	fmt.Println("Exporting PostgreSQL database to JSON...")

	// Initialize database connection
	if err := InitDB(); err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		return
	}
	defer CloseDB()

	// Get all commands (empty pattern returns all)
	jsonData, err := SearchCommands("", "json")
	if err != nil {
		fmt.Printf("Error exporting data: %v\n", err)
		return
	}

	// Determine output file path
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}

	outputFile := filepath.Join(currentDir, filename)

	// Write JSON to file
	err = os.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	fmt.Printf("Successfully exported database to: %s\n", outputFile)
	fmt.Println()
}

// replaceLast receives (string in, old string, new string) it replaces the last instance of the old string and return updated string
func replaceLast(x, y, z string) (x2 string) {
	i := strings.LastIndex(x, y)
	if i == -1 {
		return x
	}
	return x[:i] + z + x[i+len(y):]
}

// WriteLogToFile is a function that takes a message string and a file name and stores the information into the file
func WriteLogToFile(logFile, message string) {
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	logger := log.New(file, "", log.LstdFlags)
	logger.Println(": " + message)
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
