package main

import (
	"fmt"

	"github.com/gcclinux/tardigrade-mod"
)

// saveCMD takes the cmd and details (string) and parse it to tardigrade-mod to save it to the database.
func saveCmd(cmd, details string, db string) {
	tar := tardigrade.Tardigrade{}

	status := tar.AddField(cmd, details, "tardigrade.db")
	fmt.Println("returned: (", status, ")")
	fmt.Println()

}
