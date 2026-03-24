package search

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gcclinux/scmd/internal/database"
	"github.com/gcclinux/scmd/internal/util"
)

// RunCLISearch prints the result returned from PostgreSQL database.
func RunCLISearch(pattern string) {
	util.WriteLogToFile(util.WebLog, "CLI: "+pattern)

	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDB()

	received, err := database.SearchCommands(pattern, "json")
	if err != nil {
		log.Fatalf("Error searching commands: %v", err)
	}

	var dt []database.CommandRecord
	json.Unmarshal(received, &dt)

	util.CheckDB(received)

	for x := range dt {
		cmd := string(dt[x].Key)
		check := util.IsCode(dt[x].Key)
		if check {
			if !strings.HasSuffix(cmd, "{{end}}") {
				cmd = util.ReplaceLast(cmd, "}", "\n}")
			}
			cmd = strings.ReplaceAll(cmd, "\n\t\n\t", "\n\t\t")
			fmt.Printf("\nID:%v - %v", dt[x].Id, dt[x].Data)
			fmt.Println()
			fmt.Println(cmd)
			fmt.Println()
		} else {
			out, _ := json.MarshalIndent(&dt[x], "", "  ")
			fmt.Println(string(out))
		}
	}
}
