package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gcclinux/tardigrade-mod"
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

// copyDB function create a copy of your current saved commands database and store it in your ${HOME} drive
func copyDB(db string) {
	fmt.Println()
	tar := tardigrade.Tardigrade{}
	msg, status := tar.CreatedDBCopy(db)
	fmt.Println(msg, "(", status, ")")
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
