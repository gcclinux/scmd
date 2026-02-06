package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// download function is deprecated - data is now stored in PostgreSQL
func download() {
	fmt.Println()
	fmt.Println("Note: Database download functionality is no longer available.")
	fmt.Println("This application now uses PostgreSQL instead of tardigrade.db")
	fmt.Println()
	fmt.Println("To import data into PostgreSQL, use the CLI tools in the cli/ directory:")
	fmt.Println("  cd cli/")
	fmt.Println("  python import_to_postgres.py")
	fmt.Println()
}

// downloadFile take a url and a filename, downloads and save it locally in this case used for tardigrade.db
func downloadFile(filepath string, url string) (err error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
