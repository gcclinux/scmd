package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gcclinux/tardigrade-mod"
)

type BuildStruct struct {
	PageTitle string
	Pattern   string
	Id        int
	Key       string
	Data      string
	CmdTitle  string
	DescTitle string
	Version   string
	Return    string
	Status    string
	CmdFunc   string
	AllData   []string
	Code      []string
}

//go:embed templates
var tplFolder embed.FS // embeds the templates folder into variable tplFolder

func routes() {

	// create a WaitGroup
	wg := new(sync.WaitGroup)
	wg.Add(2) // create one go routine

	HTTP := 3333
	browser := true
	SSL := true
	CRT := ""
	KEY := ""

	count := len(os.Args)
	///log.Println("Count: ", count)

	if count == 2 {
		if os.Args[1] == "--web" {
			SSL = false
		} else {
			wrongSyntax()
			os.Exit(1)
		}
	}

	if count == 3 {
		wrongSyntax()
	}

	if count == 4 {
		if os.Args[2] == "-port" && isInt(os.Args[3]) {
			HTTP, _ = strconv.Atoi(os.Args[3])
		}
		if os.Args[1] == "--web" {
			SSL = false
		} else if SSL {
			CRT = os.Args[2]
			KEY = os.Args[3]
		} else {
			wrongSyntax()
			os.Exit(1)
		}
	}

	if count == 5 {
		if os.Args[2] == "-port" && isInt(os.Args[3]) {
			HTTP, _ = strconv.Atoi(os.Args[3])
		}
		if os.Args[4] == "-service" {
			browser = false
			if SSL {
				wrongSyntax()
				os.Exit(1)
			}
		}
	}

	if count == 6 {
		if os.Args[2] == "-port" && isInt(os.Args[3]) {
			HTTP, _ = strconv.Atoi(os.Args[3])
		}
		if os.Args[4] == "-service" {
			browser = false
		}
		if !SSL {
			wrongSyntax()
			os.Exit(1)
		} else {
			CRT = os.Args[4]
			KEY = os.Args[5]
		}
	}

	if count == 7 {
		if os.Args[2] == "-port" && isInt(os.Args[3]) {
			HTTP, _ = strconv.Atoi(os.Args[3])
		}
		if os.Args[4] == "-service" {
			browser = false
		}
		if SSL {
			CRT = os.Args[5]
			KEY = os.Args[6]
		} else {
			wrongSyntax()
			os.Exit(1)
		}
	}

	http.HandleFunc("/", HomePage)
	http.HandleFunc("/add", AddPage)
	http.HandleFunc("/help", HelpPage)

	if browser {
		if SSL {
			log.Println("Starting scmd web HTTPS UI on port", HTTP)
			openBrowser(fmt.Sprintf("https://localhost:%v", HTTP))
			err := http.ListenAndServeTLS(fmt.Sprintf(":%v", HTTP), CRT, KEY, nil)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Println("Starting scmd web HTTP UI on port", HTTP)
			openBrowser(fmt.Sprintf("http://localhost:%v", HTTP))
			err := http.ListenAndServe(fmt.Sprintf(":%v", HTTP), nil)
			if err != nil {
				log.Println(err)
			}
		}

	} else {
		go func() {
			if SSL {
				err := http.ListenAndServeTLS(fmt.Sprintf(":%v", HTTP), CRT, KEY, nil)
				if err != nil {
					log.Println(err)
				}
			} else {
				err := http.ListenAndServe(fmt.Sprintf(":%v", HTTP), nil)
				if err != nil {
					log.Println(err)
				}
			}
			wg.Done() // one goroutine finished
		}()
	}

	// wait until WaitGroup is done
	wg.Wait()
}

func HelpPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	tmpl := template.Must(template.ParseFS(tplFolder, "templates/help.html"))
	data := BuildStruct{
		PageTitle: "(SCMD)",
	}

	data.Version = Release
	sc := make([]string, 0)

	if r.Method == "GET" {
		tmpl.Execute(w, data)
	} else {
		r.ParseForm()
		var hidden = r.Form["hidden"][0]
		if hidden == "version" {
			msg, _, _ := versionRemote()
			versionCheck := versionCheck(msg)

			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "Current version status")
			sc = append(sc, versionCheck)
		} else if hidden == "upgrade" {
			msg, _, _ := versionRemote()
			versionCheck := versionCheck(msg)

			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "Upgrade checking current version")
			sc = append(sc, versionCheck)
			if !strings.Contains(versionCheck, "newer") && !strings.Contains(versionCheck, "already") {
				result := runUpgrade()

				sc = append(sc, "")
				sc = append(sc, "----------------------------------------------------------------------")
				sc = append(sc, "Current version status")
				sc = append(sc, result)
			}

		} else if hidden == "download" {

			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "Downloading the latest database")
			sc = append(sc, "")
			download()
			sc = append(sc, "Database download complete!")
		} else if hidden == "cli" {
			sc = append(sc, "")
			sc = append(sc, "CLI (Command Line Interface) | UI (User Interface) | SCMD (Search Commands)")
			sc = append(sc, "")
			sc = append(sc, "")
			sc = append(sc, "INFO: Display this help menu")
			sc = append(sc, "Command: scmd-Linux-x86_64(exe) --help")
			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "INFO: Opens the web UI with default Port: \"3333\"")
			sc = append(sc, "Command: scmd-Linux-x86_64(exe) --web")
			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "INFO: Opens the web UI with alternative Port:")
			sc = append(sc, "Command: scmd-Linux-x86_64(exe) --web -port [port]")
			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "INFO: Linux start UI without launching browser as a background service")
			sc = append(sc, "Command: screen -dmS SCMD scmd-Linux-aarch64 --web -port 3333 -services")
			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "INFO: Windows start UI without launching browser as a background service")
			sc = append(sc, "Command: START SCMD /B scmd-win-x86_64.exe --web -port 3333 -services")
			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "INFO: Show local and available scmd version")
			sc = append(sc, "Command: scmd-Linux-x86_64(exe) --version")
			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "INFO: Copy for the commands database and save it in Home folder")
			sc = append(sc, "Command: scmd-Linux-x86_64(exe) --copydb")
			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "INFO: Download latest tardigrade.db WRN: (override locally DB)")
			sc = append(sc, "Command: scmd-Linux-x86_64(exe) --download")
			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "INFO: Download and upgrade the latest version of the scmd application binary")
			sc = append(sc, "Command: scmd-Linux-x86_64(exe) --upgrade")
			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "INFO: Search command based on comma separated pattern(s)")
			sc = append(sc, "Command: scmd-Linux-x86_64(exe) --search \"pattern(s)\"")
			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "INFO: Save new command with description in the local database")
			sc = append(sc, "Command: scmd-Linux-x86_64(exe) --save \"command\" \"description\"")
			sc = append(sc, "")
		}

		data.AllData = sc
		tmpl.Execute(w, data)
	}
}

func AddPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	tar := tardigrade.Tardigrade{}
	tmpl := template.Must(template.ParseFS(tplFolder, "templates/add.html"))
	data := BuildStruct{
		PageTitle: "(SCMD)",
	}

	remoteAddr := r.RemoteAddr
	WriteLogToFile(webLog, "ADD: "+remoteAddr)

	data.Version = Release

	if r.Method == "GET" {
		tmpl.Execute(w, data)
	} else {
		r.ParseForm()
		var command = r.Form["command"][0]
		var description = r.Form["description"][0]

		WriteLogToFile(webLog, remoteAddr+" : "+command)

		save := true
		status := false
		var _, received = tar.SelectSearch(command, "json")
		bytes := received
		var dt []tardigrade.MyStruct
		json.Unmarshal(bytes, &dt)

		checkDB(received)

		for x := range dt {
			cmd := string(dt[x].Key)
			check := strings.Contains(command, cmd)
			if check {
				save = false
			}
		}

		if save {
			status = tar.AddField(command, description)
			data.Status = fmt.Sprintf("%t", status)
		} else {
			data.Status = fmt.Sprintf("%v", "(false) Duplicate command!")
		}

		data.Return = "Return Status: "
		data.DescTitle = "Description: "
		data.Data = fmt.Sprintf("%v", description)
		data.CmdTitle = "Command:  "
		data.Key = fmt.Sprintf("%v", command)

		tmpl.Execute(w, data)
	}
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	tar := tardigrade.Tardigrade{}
	tmpl := template.Must(template.ParseFS(tplFolder, "templates/home.html"))

	remoteAddr := r.RemoteAddr
	WriteLogToFile(webLog, "HOME: "+remoteAddr)

	data := BuildStruct{
		PageTitle: "(SCMD)",
	}

	data.Version = Release
	sc := make([]string, 0)
	scode := make([]string, 0)

	if r.Method == "GET" {
		tmpl.Execute(w, data)
	} else {
		r.ParseForm()
		var pattern = r.Form["pattern"][0]

		if len(pattern) < 3 {
			tmpl.Execute(w, data)
		} else {
			WriteLogToFile(webLog, "SEARCH: "+pattern)

			var _, received = tar.SelectSearch(pattern, "raw")
			bytes := received
			var dt []BuildStruct
			json.Unmarshal(bytes, &dt)

			data.Pattern = checkDB(received)

			for x := range dt {

				code := isCode(dt[x].Key)

				if code {
					funccmd := dt[x].Key
					if !strings.HasSuffix(funccmd, "{{end}}") {
						funccmd = replaceLast(funccmd, "}", "\n}")
					}
					funccmd = strings.ReplaceAll(funccmd, "\n\t\n\t", "\n\n\t")
					scode = append(scode, "//ID: "+strconv.Itoa(dt[x].Id)+" - "+dt[x].Data)
					scode = append(scode, funccmd)
				} else {
					sc = append(sc, "----------------------------------------------------------------------")
					sc = append(sc, "# ID: ")
					sc = append(sc, strconv.Itoa(dt[x].Id))
					sc = append(sc, "# Description: ")
					sc = append(sc, fmt.Sprintf("\"%s\"", string(dt[x].Data)))
					sc = append(sc, "# Command : ")
					sc = append(sc, string(dt[x].Key))
					sc = append(sc, "")
				}
			}

			data.AllData = sc
			data.Code = scode
			tmpl.Execute(w, data)
		}
	}
}
