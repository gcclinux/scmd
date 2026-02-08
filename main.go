package main

// Updated Wed  8 Mar 2023 15:46:28 GMT

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// content holds our static web server content.
func main() {
	LoadEnv()
	msg, _, _ := versionRemote()
	count := len(os.Args)

	if count == 2 || count == 1 {
		if count == 1 {
			helpHere(getName())
			os.Exit(0)
		}
		arg1 := os.Args[1]
		if arg1 == "--help" {
			helpHere(getName())
		} else if arg1 == "--version" {
			versionCheck(msg)
		} else if arg1 == "--search" {
			fmt.Println("Usage: \t", getName(), "--search [pattern(s)]")
		} else if arg1 == "--download" {
			download()
		} else if arg1 == "--upgrade" {
			runUpgrade()
		} else if arg1 == "--web" {
			routes()
		} else if arg1 == "--ssl" {
			routes()
		} else if arg1 == "--interactive" || arg1 == "-i" || arg1 == "--cli" {
			StartInteractiveMode()
		} else if arg1 == "--generate-embeddings" {
			// Initialize providers
			InitGemini()
			InitOllama()
			// Initialize database
			if err := InitDB(); err != nil {
				fmt.Printf("Failed to connect to database: %v\n", err)
				os.Exit(1)
			}
			defer CloseDB()
			// Generate embeddings
			if err := GenerateEmbeddingsForAll(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		} else if arg1 == "--embedding-stats" {
			// Initialize database
			if err := InitDB(); err != nil {
				fmt.Printf("Failed to connect to database: %v\n", err)
				os.Exit(1)
			}
			defer CloseDB()
			// Show stats
			if err := CheckEmbeddingStats(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			wrongSyntax()
		}
	} else if count == 3 {
		if (os.Args[1]) == "--search" {
			search(os.Args[2])
		} else if (os.Args[1]) == "--web" && (os.Args[2]) == "-block" {
			routes()
		} else if os.Args[1] == "--copydb" {
			copyDB(os.Args[2])
		} else if os.Args[1] == "--save" {
			// Check if there is data in stdin
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				stdin, err := io.ReadAll(os.Stdin)
				if err != nil {
					fmt.Println("Error reading from stdin:", err)
					return
				}
				// When piping, the Stdin content is the Command (Key),
				// and the argument provided is the Description (Data/Details).
				saveCmd(string(stdin), os.Args[2])
			} else {
				fmt.Println("Usage: scmd --save [command] [description]")
				fmt.Println("To save a large script/command: scmd --save [description] < script.sh")
			}
		} else {
			wrongSyntax()
		}
	} else if count == 4 {
		if os.Args[1] == "--web" {
			routes()
		} else if os.Args[1] == "--ssl" {
			routes()
		} else if os.Args[1] == "--save" {
			saveCmd(os.Args[2], os.Args[3])
		} else {
			wrongSyntax()
		}
	} else if count == 5 {
		if (os.Args[1]) == "--save" {
			saveCmd(os.Args[2], os.Args[3])
		} else {
			routes()
		}

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
