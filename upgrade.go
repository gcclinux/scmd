package main

// Updated Mon  6 Mar 18:34:03 GMT 2023

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// runUpgrade function will check release note for current version and then upgrade if available.
// upgrades will only work for compiled binaries available in https://github.com/gcclinux/scmd/raw/main/bin/
func runUpgrade() string {

	path := string(os.PathSeparator)
	file := getFile()
	remote_version := ""

	local_int := strings.Replace(Release, ".", "", -1)
	msg, remote_version, upgrade := versionRemote()

	if upgrade {

		local := fmt.Sprintf("%v%v%v", filepath.Dir(os.Args[0]), string(path), filepath.Base(os.Args[0]))
		old := fmt.Sprintf("%v%v%v%v%v%v", filepath.Dir(os.Args[0]), string(path), "v", local_int, "-", filepath.Base(os.Args[0]))
		remote := fmt.Sprintf("%v%v", "https://github.com/gcclinux/scmd/releases/download/latest/", file)

		replaceFile(local, old)
		downloadFile(file, remote)

		err := os.Chmod(local, 0750)
		if err != nil {
			log.Fatal(err)
		}

		msg = fmt.Sprintf("Upgraded %v (%v) to latest version (%v)\n\r", file, Release, remote_version)
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
	FILE_NAME := ""
	if runtime.GOOS == "windows" && runtime.GOARCH == "amd64" {
		FILE_NAME = "scmd-win-x86_64.exe"
	} else if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		FILE_NAME = "scmd-Linux-x86_64"
	} else if runtime.GOOS == "linux" && runtime.GOARCH == "arm64" {
		FILE_NAME = "scmd-Linux-aarch64"
	} else if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		FILE_NAME = "scmd-Darwin-arm64"
	} else {
		log.Println("unknown")
	}
	return FILE_NAME
}
