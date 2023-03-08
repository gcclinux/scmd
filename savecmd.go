package main

import (
	"fmt"

	"github.com/gcclinux/tardigrade-mod"
)

// saveCMD takes the cmd and details 9string) and parse it to tardigrade-mod to save it to the database.
func saveCmd(cmd, details string) {
	tar := tardigrade.Tardigrade{}
	fmt.Println()
	status := tar.AddField(cmd, details)
	fmt.Println("returned: (", status, ")")
	fmt.Println()

}
