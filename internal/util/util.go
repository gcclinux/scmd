package util

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

const WebLog = "scmdweb.log"

// CheckDB will see if the Database exist and if it contains any data.
func CheckDB(received []byte) string {
	s := string(received)
	if (strings.Contains(s, "Database") && strings.Contains(s, "missing")) ||
		(strings.Contains(s, "Database") && strings.Contains(s, "empty")) {
		fmt.Println(s)
		return s
	}
	return ""
}

// IsInt checks if a string is a valid integer.
func IsInt(in string) bool {
	_, err := strconv.ParseInt(in, 10, 0)
	return err == nil
}

// IsCode detects if a string contains code patterns.
func IsCode(in string) bool {
	if strings.Contains(in, "func") && strings.Contains(in, "{") {
		return true
	}
	if strings.Contains(in, "class") && strings.Contains(in, "{") && !strings.Contains(in, "PS>") {
		return true
	}
	if strings.Contains(in, "println") && strings.Contains(in, "{") {
		return true
	}
	if strings.Contains(in, "public") && strings.Contains(in, "{") {
		return true
	}
	if strings.Contains(in, "<html>") && strings.Contains(in, "<body>") {
		return true
	}
	if strings.Contains(in, "<script>") && strings.Contains(in, "</script>") {
		return true
	}
	if strings.Contains(in, "stdio.h") && strings.Contains(in, "scanf") {
		return true
	}
	if strings.Contains(in, "##instructions##") {
		return true
	}
	if strings.Contains(in, "{{end}}") {
		return true
	}
	if strings.Contains(in, "#") && strings.Contains(in, "/usr/bin/python") {
		return true
	}
	if strings.Contains(in, "import") && strings.Contains(in, "{") {
		return true
	}
	return false
}

// ReplaceLast replaces the last occurrence of old with new in x.
func ReplaceLast(x, y, z string) string {
	i := strings.LastIndex(x, y)
	if i == -1 {
		return x
	}
	return x[:i] + z + x[i+len(y):]
}

// WriteLogToFile writes a message to a log file.
func WriteLogToFile(logFile, message string) {
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	logger := log.New(file, "", log.LstdFlags)
	logger.Println(": " + message)
}

// GetOutboundIP returns the local outbound IP address.
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

// OpenBrowser launches the default browser depending on the OS.
func OpenBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

// CopyDB exports all commands from PostgreSQL to a JSON file.
// searchFn should be a function that returns all commands as JSON (e.g., database.SearchCommands("", "json")).
func CopyDB(filename string, searchFn func(string, string) ([]byte, error)) {
	fmt.Println()
	fmt.Println("Exporting PostgreSQL database to JSON...")

	jsonData, err := searchFn("", "json")
	if err != nil {
		fmt.Printf("Error exporting data: %v\n", err)
		return
	}

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}

	outputFile := currentDir + "/" + filename
	err = os.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	fmt.Printf("Successfully exported database to: %s\n", outputFile)
	fmt.Println()
}
