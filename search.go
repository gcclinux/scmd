package main

import (
	"encoding/json"
	"fmt"

	"github.com/gcclinux/tardigrade-mod"
)

// search prints the result returned by the Tardigrade mod
func search(pattern string) {
	tar := tardigrade.Tardigrade{}
	var _, received = tar.SelectSearch(pattern, "json")
	bytes := received
	var dt []tardigrade.MyStruct
	json.Unmarshal(bytes, &dt)

	checkDB(received)

	for x := range dt {
		out, _ := tar.MyIndent(&dt[x], "", "  ")
		fmt.Printf("%v", string(out))
		fmt.Println()
	}
}
