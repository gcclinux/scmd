package main

// Updated Wed  8 Mar 2023 15:46:28 GMT

import (
	"fmt"
	"os"
	"path/filepath"
)

// content holds our static web server content.
func main() {

	msg, _, _ := versionRemote()

	if len(os.Args) == 2 {
		arg1 := os.Args[1]
		if arg1 == "--help" {
			helpHere(getName())
		} else if arg1 == "--version" {
			versionCheck(msg)
		} else if arg1 == "--search" {
			fmt.Println("Usage: \t", getName(), "--search [pattern(s)]")
		} else if arg1 == "--download" {
			download()
		} else if arg1 == "--copydb" {
			copyDB()
		} else if arg1 == "--upgrade" {
			runUpgrade()
		} else if arg1 == "--web" {
			routes()
		} else {
			wrongSyntax()
		}
	} else if len(os.Args) == 3 {
		if (os.Args[1]) == "--search" {
			search(os.Args[2])
		} else {
			wrongSyntax()
		}
	} else if len(os.Args) == 4 {
		if (os.Args[1]) == "--save" {
			saveCmd(os.Args[2], os.Args[3])
		} else if os.Args[1] == "--web" {
			routes()
		} else {
			wrongSyntax()
		}
	} else if len(os.Args) == 5 {
		if os.Args[1] == "--web" {
			routes()
		}
	} else {
		wrongSyntax()
	}
}

func getName() string {
	return filepath.Base(os.Args[0])
}
