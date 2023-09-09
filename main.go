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
	count := len(os.Args)

	if count == 2 {
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
		} else if arg1 == "--ssl" {
			routes()
		} else {
			wrongSyntax()
		}
	} else if count == 3 {
		if (os.Args[1]) == "--search" {
			search(os.Args[2])
		} else if (os.Args[1]) == "--web" && (os.Args[2]) == "-block" {
			routes()
		} else {
			wrongSyntax()
		}
	} else if count == 4 {
		if (os.Args[1]) == "--save" {
			saveCmd(os.Args[2], os.Args[3])
		} else if os.Args[1] == "--web" {
			routes()
		} else if os.Args[1] == "--ssl" {
			routes()
		} else {
			wrongSyntax()
		}
	} else if count == 5 {
		routes()
	} else if count == 6 && os.Args[count-1] != "-block" {
		if os.Args[1] == "--ssl" {
			routes()
		} else {
			wrongSyntax()
		}
	} else if count == 6 && os.Args[count-1] == "-block" && os.Args[1] == "--web" || os.Args[1] == "--ssl" {
		routes()
	} else if len(os.Args) == 7 || len(os.Args) == 8 || len(os.Args) == 9 {
		if os.Args[1] == "--ssl" {
			routes()
		} else {
			wrongSyntax()
		}
	} else {
		wrongSyntax()
	}
}

func getName() string {
	return filepath.Base(os.Args[0])
}
