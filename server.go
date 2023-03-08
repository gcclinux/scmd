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
	AllData   []string
}

//go:embed templates
var tplFolder embed.FS // embeds the templates folder into variable tplFolder

func routes() {

	// create a WaitGroup
	wg := new(sync.WaitGroup)
	wg.Add(2) // create one go routine

	HTTP := 3333
	browser := true

	count := len(os.Args)

	if count == 4 {
		if os.Args[2] == "-port" && isInt(os.Args[3]) {
			HTTP, _ = strconv.Atoi(os.Args[3])
		}
	} else if count == 5 {
		if os.Args[2] == "-port" && isInt(os.Args[3]) {
			HTTP, _ = strconv.Atoi(os.Args[3])
		}
		if os.Args[4] == "-service" {
			browser = false
		} else {
			log.Println("Incorrect syntax: (", os.Args[4], ") is not an option")
			os.Exit(0)
		}
	}

	http.HandleFunc("/", HomePage)
	http.HandleFunc("/add", AddPage)

	if browser {
		log.Println("Starting scmd web UI on port", HTTP)
		openBrowser(fmt.Sprintf("http://localhost:%v", HTTP))
		err := http.ListenAndServe(fmt.Sprintf(":%v", HTTP), nil)
		if err != nil {
			log.Println(err)
		}
	} else {
		go func() {
			err := http.ListenAndServe(fmt.Sprintf(":%v", HTTP), nil)
			if err != nil {
				log.Println(err)
			}
			wg.Done() // one goroutine finished
		}()

		go func() {
			log.Println("Starting scmd web service on port", HTTP)
			wg.Done()
		}()
	}

	// wait until WaitGroup is done
	wg.Wait()
}
func AddPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	tar := tardigrade.Tardigrade{}
	tmpl := template.Must(template.ParseFS(tplFolder, "templates/add.html"))
	data := BuildStruct{
		PageTitle: "(SCMD)",
	}

	data.Version = Release

	if r.Method == "GET" {
		tmpl.Execute(w, data)
	} else {
		r.ParseForm()
		var command = r.Form["command"][0]
		var description = r.Form["description"][0]

		status := tar.AddField(command, description)
		data.Status = fmt.Sprintf("%t", status)
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

	data := BuildStruct{
		PageTitle: "(SCMD)",
	}

	data.Version = Release
	sc := make([]string, 0)

	if r.Method == "GET" {
		tmpl.Execute(w, data)
	} else {
		r.ParseForm()
		var pattern = r.Form["pattern"][0]
		var _, received = tar.SelectSearch(pattern, "raw")
		bytes := received
		var dt []BuildStruct
		json.Unmarshal(bytes, &dt)

		data.Pattern = checkDB(received)

		for x := range dt {
			fmt.Printf("id: %d, key: %v, data: %s\n", dt[x].Id, dt[x].Key, dt[x].Data)
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "# Description:")
			sc = append(sc, fmt.Sprintf("\"%s\"", string(dt[x].Data)))
			sc = append(sc, "# Command : ")
			sc = append(sc, string(dt[x].Key))
			sc = append(sc, "")
		}

		data.AllData = sc
		tmpl.Execute(w, data)

	}
}
