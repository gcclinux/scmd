package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// download function is the caller to download the online tardigrade.db
func download() {

	file := "tardigrade.db"
	remote := "https://raw.githubusercontent.com/gcclinux/scmd/main/tardigrade.db"

	downloadFile(file, remote)

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
