package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gcclinux/tardigrade-mod"
)

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
func copyDB() {
	fmt.Println()
	tar := tardigrade.Tardigrade{}
	msg, status := tar.CreatedDBCopy()
	fmt.Println(msg, "(", status, ")")
	fmt.Println()
}
