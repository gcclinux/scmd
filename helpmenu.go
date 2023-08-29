package main

import (
	"fmt"
)

const (
	InfoColor    = "\033[1;34m%s\033[0m"
	NoticeColor  = "\033[1;36m%s\033[0m"
	WarningColor = "\033[1;33m%s\033[0m"
	ErrorColor   = "\033[1;31m%s\033[0m"
	DebugColor   = "\033[0;36m%s\033[0m"
)

func helpHere(name string) {
	fmt.Println()
	fmt.Println("Version:", Release)
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Display this help menu\n\r")
	fmt.Println("Usage: \t", name, "--help")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Opens the Web UI with default Port: \"3333\" \n\r")
	fmt.Println("Usage: \t", name, "--web")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Opens the Web UI with alternative Port:\n\r")
	fmt.Println("Usage: \t", name, "--web", "-port", "[port]")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Starts SCMD without launching Web UI\n\r")
	fmt.Println("Usage: \t", name, "--web", "-port", "[port]", "-service")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Opens SSL Web UI with default Port: \"3333\" \n\r")
	fmt.Println("Usage: \t", name, "--ssl", "[certificate.pem]", "[privkey.pem]")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Opens SSL web UI with alternative Port:\n\r")
	fmt.Println("Usage: \t", name, "--ssl", "-port", "[port]", "[certificate.pem]", "[privkey.pem]")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Starts SCMD SSL without launching Web UI\n\r")
	fmt.Println("Usage: \t", name, "--ssl", "-port", "[port]", "-service", "[certificate.pem]", "[privkey.pem]")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Show local and available scmd version\n\r")
	fmt.Println("Usage: \t", name, "--version")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Create a copy for the commands database and save it in Home folder\n\r")
	fmt.Println("Usage: \t", name, "--copydb")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Download all available commands database from online (override locally saved commands)\n\r")
	fmt.Println("Usage: \t", name, "--download")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Download and upgrade the latest version of the scmd application binary\n\r")
	fmt.Println("Usage: \t", name, "--upgrade")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Search command based on comma separated pattern(s)\n\r")
	fmt.Println("Usage: \t", name, "--search [pattern(s)]")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Save new command with description in the local database\n\r")
	fmt.Println("Usage: \t", name, "--save [command] [description]")
	fmt.Println()
}

func wrongSyntax() {
	fmt.Println()
	fmt.Println("Usage: \t", getName(), "--help")
	fmt.Println()
}
