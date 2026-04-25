package updater

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// RunUpgrade checks for a new version and upgrades the binary if available.
func RunUpgrade() string {
	path := string(os.PathSeparator)
	file := getFile()

	localInt := strings.Replace(Release, ".", "", -1)
	msg, remoteVersion, upgrade := VersionRemote()

	if upgrade {
		local := fmt.Sprintf("%v%v%v", filepath.Dir(os.Args[0]), string(path), filepath.Base(os.Args[0]))
		old := fmt.Sprintf("%v%v%v%v%v%v", filepath.Dir(os.Args[0]), string(path), "v", localInt, "-", filepath.Base(os.Args[0]))
		// Ensure the version string starts with 'v' for the GitHub release tag
		versionTag := remoteVersion
		if !strings.HasPrefix(versionTag, "v") {
			versionTag = "v" + versionTag
		}
		
		remote := fmt.Sprintf("https://github.com/gcclinux/scmd/releases/download/%v/%v", versionTag, file)

		replaceFile(local, old)
		
		fmt.Printf("Downloading new version from %s...\n", remote)
		err := DownloadFile(local, remote)
		if err != nil {
			log.Fatalf("Failed to download new version: %v", err)
		}

		err = os.Chmod(local, 0750)
		if err != nil {
			log.Fatal(err)
		}

		msg = fmt.Sprintf("Upgraded %v (%v) to latest version (%v)\n\r", file, Release, remoteVersion)
	}
	fmt.Println()
	fmt.Println(msg)

	return msg
}

func replaceFile(file string, local string) {
	err := os.Rename(file, local)
	if err != nil {
		log.Fatal(err)
	}
}

func getFile() string {
	fileName := ""
	if runtime.GOOS == "windows" && runtime.GOARCH == "amd64" {
		fileName = "scmd-windows-amd64.exe"
	} else if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		fileName = "scmd-linux-amd64"
	} else if runtime.GOOS == "linux" && runtime.GOARCH == "arm64" {
		fileName = "scmd-linux-aarch64"
	} else if runtime.GOOS == "darwin" && runtime.GOARCH == "amd64" {
		fileName = "scmd-darwin-amd64"
	} else if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		fileName = "scmd-darwin-arm64"
	} else {
		log.Println("unknown architecture/os")
	}
	return fileName
}

// Download prints the deprecation notice for the old download functionality.
func Download() {
	fmt.Println()
	fmt.Println("Note: Database download functionality is no longer available.")
	fmt.Println("This application uses an automated SQLite database configuration.")
	fmt.Println()
}

// DownloadFile downloads a file from a URL and saves it locally.
func DownloadFile(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}
