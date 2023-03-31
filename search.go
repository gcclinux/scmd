package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gcclinux/tardigrade-mod"
)

// search prints the result returned by the Tardigrade mod
func search(pattern string) {

	WriteLogToFile("scmd.log", pattern)

	tar := tardigrade.Tardigrade{}

	var _, received = tar.SelectSearch(pattern, "json")
	bytes := received
	var dt []tardigrade.MyStruct
	json.Unmarshal(bytes, &dt)

	checkDB(received)

	for x := range dt {
		out, _ := tar.MyIndent(&dt[x], "", "  ")
		cmd := string(dt[x].Key)
		check := isCode(dt[x].Key)
		if check {
			if !strings.HasSuffix(cmd, "{{end}}") {
				cmd = replaceLast(cmd, "}", "\n}")
			}
			cmd = strings.ReplaceAll(cmd, "\n\t\n\t", "\n\t\t")
			fmt.Printf("\nID:%v - %v", dt[x].Id, dt[x].Data)
			fmt.Println()
			fmt.Println(cmd)
			fmt.Println()
		} else {
			fmt.Println(string(out))
		}
	}
}
