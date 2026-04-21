package updater

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Release is the current application version.
const Release = "2.0.7"

// VersionCheck displays the local and remote version.
func VersionCheck(msg string) string {
	fmt.Println()
	fmt.Println(msg)
	fmt.Println()
	return msg
}

// VersionRemote retrieves the remote version and compares it.
func VersionRemote() (msg string, remote string, upgrade bool) {
	upgrade = false
	remoteVersion := ""
	file := filepath.Base(os.Args[0])

	lines, err := urlToLines("https://raw.githubusercontent.com/gcclinux/scmd/main/release")
	if err != nil {
		log.Println(err)
	}

	for _, line := range lines {
		remoteVersion = line
	}

	localInt := strings.Replace(Release, ".", "", -1)
	remoteInt := strings.Replace(remoteVersion, ".", "", -1)

	if localInt == remoteInt {
		msg = fmt.Sprintf("Currently \"%v\" is already at the latest version (%v)", file, remoteVersion)
	} else if localInt > remoteInt {
		msg = fmt.Sprintf("Currently \"%v\" (%v) is newer than the online version (%v)", file, Release, remoteVersion)
	} else if localInt < remoteInt {
		msg = fmt.Sprintf("Newer version for \"%v\" (%v) is available for downloading v(%v)", file, Release, remoteVersion)
		upgrade = true
	}

	return msg, remoteVersion, upgrade
}

func urlToLines(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return linesFromReader(resp.Body)
}

func linesFromReader(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
