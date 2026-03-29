package main

import (
	"embed"
	"fmt"
	"io"
	"os"

	"github.com/gcclinux/scmd/internal/ai"
	"github.com/gcclinux/scmd/internal/cli"
	"github.com/gcclinux/scmd/internal/config"
	"github.com/gcclinux/scmd/internal/database"
	"github.com/gcclinux/scmd/internal/markdown"
	"github.com/gcclinux/scmd/internal/search"
	"github.com/gcclinux/scmd/internal/server"
	"github.com/gcclinux/scmd/internal/setup"
	"github.com/gcclinux/scmd/internal/updater"
	"github.com/gcclinux/scmd/internal/util"
)

//go:embed templates
var tplFolder embed.FS

func main() {
	config.LoadConfig()
	msg, _, _ := updater.VersionRemote()
	count := len(os.Args)

	// Pass embedded templates to the server package
	server.SetTemplates(tplFolder)

	if count == 2 || count == 1 {
		if count == 1 {
			cli.PrintHelp(cli.GetName())
			os.Exit(0)
		}
		arg1 := os.Args[1]
		if arg1 == "--help" {
			cli.PrintHelp(cli.GetName())
		} else if arg1 == "--version" {
			updater.VersionCheck(msg)
		} else if arg1 == "--search" {
			fmt.Println("Usage: \t", cli.GetName(), "--search [pattern(s)]")
		} else if arg1 == "--download" {
			updater.Download()
		} else if arg1 == "--upgrade" {
			updater.RunUpgrade()
		} else if arg1 == "--create-db" {
			database.SetupDatabase()
		} else if arg1 == "--create-db-postgresql" {
			setup.SetupPostgreSQL()
		} else if arg1 == "--create-db-sqlite" {
			setup.SetupSQLite()
		} else if arg1 == "--connect-db-sqlite" {
			setup.ConnectSQLite()
		} else if arg1 == "--connect-db-postgresql" {
			setup.ConnectPostgreSQL()
		} else if arg1 == "--server-ollama" {
			setup.SetupOllama()
		} else if arg1 == "--server-gemini" {
			setup.SetupGemini()
		} else if arg1 == "--web" {
			server.Routes()
		} else if arg1 == "--ssl" {
			server.Routes()
		} else if arg1 == "--interactive" || arg1 == "-i" || arg1 == "--cli" {
			cli.StartInteractiveMode()
		} else if arg1 == "--generate-embeddings" {
			ai.InitProviders()
			if err := database.InitDB(); err != nil {
				fmt.Printf("Failed to connect to database: %v\n", err)
				os.Exit(1)
			}
			defer database.CloseDB()
			if err := ai.GenerateEmbeddingsForAll(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		} else if arg1 == "--embedding-stats" {
			if err := database.InitDB(); err != nil {
				fmt.Printf("Failed to connect to database: %v\n", err)
				os.Exit(1)
			}
			defer database.CloseDB()
			if err := ai.CheckEmbeddingStats(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		} else if arg1 == "--copydb" {
			if err := database.InitDB(); err != nil {
				fmt.Printf("Failed to connect to database: %v\n", err)
				os.Exit(1)
			}
			defer database.CloseDB()
			util.CopyDB("scmd_export.json", database.SearchCommands)
		} else {
			cli.PrintWrongSyntax()
		}
	} else if count == 3 {
		if os.Args[1] == "--search" {
			search.RunCLISearch(os.Args[2])
		} else if os.Args[1] == "--import" {
			ai.InitProviders()
			if err := database.InitDB(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
				os.Exit(1)
			}
			defer database.CloseDB()
			title, err := markdown.ImportMarkdown(os.Args[2], ai.GetBestEmbedding)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Document imported successfully: %s\n", title)
		} else if os.Args[1] == "--web" && os.Args[2] == "-block" {
			server.Routes()
		} else if os.Args[1] == "--copydb" {
			if err := database.InitDB(); err != nil {
				fmt.Printf("Failed to connect to database: %v\n", err)
				os.Exit(1)
			}
			defer database.CloseDB()
			util.CopyDB(os.Args[2], database.SearchCommands)
		} else if os.Args[1] == "--save" {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				stdin, err := io.ReadAll(os.Stdin)
				if err != nil {
					fmt.Println("Error reading from stdin:", err)
					return
				}
				cli.SaveCmd(string(stdin), os.Args[2])
			} else {
				fmt.Println("Usage: scmd --save [command] [description]")
				fmt.Println("To save a large script/command: scmd --save [description] < script.sh")
			}
		} else {
			cli.PrintWrongSyntax()
		}
	} else if count == 4 {
		if os.Args[1] == "--web" {
			server.Routes()
		} else if os.Args[1] == "--ssl" {
			server.Routes()
		} else if os.Args[1] == "--save" {
			cli.SaveCmd(os.Args[2], os.Args[3])
		} else {
			cli.PrintWrongSyntax()
		}
	} else if count == 5 {
		if os.Args[1] == "--save" {
			cli.SaveCmd(os.Args[2], os.Args[3])
		} else {
			server.Routes()
		}
	} else if count == 6 && os.Args[count-1] != "-block" {
		if os.Args[1] == "--ssl" {
			server.Routes()
		} else {
			cli.PrintWrongSyntax()
		}
	} else if count == 6 && os.Args[count-1] == "-block" && (os.Args[1] == "--web" || os.Args[1] == "--ssl") {
		server.Routes()
	} else if len(os.Args) == 7 || len(os.Args) == 8 || len(os.Args) == 9 {
		if os.Args[1] == "--ssl" {
			server.Routes()
		} else {
			cli.PrintWrongSyntax()
		}
	} else {
		cli.PrintWrongSyntax()
	}
}
