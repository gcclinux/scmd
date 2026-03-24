package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gcclinux/scmd/internal/ai"
	"github.com/gcclinux/scmd/internal/database"
	"github.com/gcclinux/scmd/internal/updater"
	"github.com/gcclinux/scmd/internal/util"
)

func helpPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	tmpl := template.Must(template.ParseFS(tplFolder, "templates/help.html"))

	data := BuildStruct{
		PageTitle: "(HELP)",
	}

	if os.Args[len(os.Args)-1] == "-block" {
		data.Insert = false
	} else {
		data.Insert = true
	}

	data.Version = updater.Release
	sc := make([]string, 0)

	if r.Method == "GET" {
		tmpl.Execute(w, data)
	} else {
		r.ParseForm()
		var hidden = r.Form["hidden"][0]
		if hidden == "version" {
			msg, _, _ := updater.VersionRemote()
			vCheck := updater.VersionCheck(msg)
			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "Current version status")
			sc = append(sc, vCheck)
		} else if hidden == "upgrade" {
			msg, _, _ := updater.VersionRemote()
			vCheck := updater.VersionCheck(msg)
			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "Upgrade checking current version")
			sc = append(sc, vCheck)
			if !strings.Contains(vCheck, "newer") && !strings.Contains(vCheck, "already") {
				result := updater.RunUpgrade()
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
			updater.Download()
			sc = append(sc, "Database download complete!")
		} else if hidden == "cli" {
			sc = append(sc, "")
			sc = append(sc, "CLI (Command Line Interface) | UI (User Interface) | SCMD (Search Commands)")
			sc = append(sc, "")
			sc = append(sc, "")
			sc = append(sc, "INFO: Display this help menu")
			sc = append(sc, "Command: scmd --help")
			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "INFO: Start Interactive CLI Mode")
			sc = append(sc, "Command: scmd --cli")
			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "INFO: Opens the web UI with default Port: \"3333\"")
			sc = append(sc, "Command: scmd --web")
			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "INFO: Show local and available scmd version")
			sc = append(sc, "Command: scmd --version")
			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "INFO: Search command based on comma separated pattern(s)")
			sc = append(sc, "Command: scmd --search [pattern(s)]")
			sc = append(sc, "")
			sc = append(sc, "----------------------------------------------------------------------")
			sc = append(sc, "INFO: Save new command with description in the local database")
			sc = append(sc, "Command: scmd --save [command] [description]")
			sc = append(sc, "")
		}

		data.AllData = sc
		tmpl.Execute(w, data)
	}
}

func addPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	tmpl := template.Must(template.ParseFS(tplFolder, "templates/add.html"))
	data := BuildStruct{
		PageTitle: "(SCMD)",
	}

	remoteAddr := r.RemoteAddr
	util.WriteLogToFile(util.WebLog, "ADD: "+remoteAddr)
	data.Version = updater.Release

	if r.Method == "GET" {
		tmpl.Execute(w, data)
	} else {
		r.ParseForm()
		var command = r.Form["command"][0]
		var description = r.Form["description"][0]

		util.WriteLogToFile(util.WebLog, remoteAddr+" : "+command)

		save := true
		status := false

		exists, err := database.CheckCommandExists(command)
		if err != nil {
			log.Printf("Error checking command existence: %v", err)
			data.Status = "(false) Error checking database!"
		} else if exists {
			save = false
			data.Status = "(false) Duplicate command!"
		}

		if save {
			success, err := database.AddCommand(command, description, ai.GetBestEmbedding)
			if err != nil {
				log.Printf("Error adding command: %v", err)
				data.Status = "(false) Error saving command!"
			} else {
				status = success
				data.Status = fmt.Sprintf("%t", status)
			}
		}

		data.Return = "Return Status: "
		data.DescTitle = "Description: "
		data.Data = fmt.Sprintf("%v", description)
		data.CmdTitle = "Command:  "
		data.Key = fmt.Sprintf("%v", command)

		tmpl.Execute(w, data)
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	tmpl := template.Must(template.ParseFS(tplFolder, "templates/home.html"))

	remoteAddr := r.RemoteAddr
	util.WriteLogToFile(util.WebLog, "HOME: "+remoteAddr)

	data := BuildStruct{
		PageTitle: "(SCMD)",
	}

	if os.Args[len(os.Args)-1] == "-block" {
		data.Insert = false
	} else {
		data.Insert = true
	}

	data.Version = updater.Release
	data.AIProviderLabel = ai.GetProviderLabel()

	if r.Method == "GET" {
		tmpl.Execute(w, data)
	} else {
		r.ParseForm()
		var pattern = r.Form["pattern"][0]

		if len(pattern) < 3 {
			tmpl.Execute(w, data)
		} else {
			util.WriteLogToFile(util.WebLog, "SEARCH: "+pattern)

			results, aiResponse, aiTokens, err := ai.SmartSearch(pattern, true)
			if err != nil {
				log.Printf("Error searching commands: %v", err)
				data.Pattern = "Error searching database"
				tmpl.Execute(w, data)
				return
			}

			var pages []string

			if aiResponse != "" {
				tokenStr := "Usage Not Tracked"
				if aiTokens > 0 {
					tokenStr = strconv.Itoa(aiTokens)
				}
				aiPage := fmt.Sprintf("## AI-Generated Response\n\n**TOKEN:** %s\n\n%s", tokenStr, aiResponse)
				pages = append(pages, aiPage)
			}

			if len(results) == 0 && aiResponse == "" {
				data.Pattern = "No matches found"
			}

			for _, record := range results {
				code := util.IsCode(record.Key)
				var cmdFormatted string

				if code {
					funccmd := record.Key
					if !strings.HasSuffix(funccmd, "{{end}}") {
						funccmd = util.ReplaceLast(funccmd, "}", "\n}")
					}
					funccmd = strings.ReplaceAll(funccmd, "\n\t\n\t", "\n\n\t")
					cmdFormatted = fmt.Sprintf("```go\n%s\n```", funccmd)
				} else {
					cmd := record.Key
					if strings.Contains(cmd, "```") || strings.Contains(cmd, "\n") {
						if strings.Contains(cmd, "```") {
							cmdFormatted = cmd
						} else {
							cmdFormatted = fmt.Sprintf("```\n%s\n```", cmd)
						}
					} else {
						cmdFormatted = fmt.Sprintf("```\n%s\n```", cmd)
					}
				}

				pageContent := fmt.Sprintf("## ID: %d Description\n%s\n\n## Command\n\n%s", record.Id, record.Data, cmdFormatted)
				pages = append(pages, pageContent)
			}

			data.Pages = pages
			data.PageQuery = pattern
			data.SaveStatus = ""
			tmpl.Execute(w, data)
		}
	}
}

func gamePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	tmpl := template.Must(template.ParseFS(tplFolder, "templates/game.html"))

	remoteAddr := r.RemoteAddr
	util.WriteLogToFile(util.WebLog, "GAME: "+remoteAddr)

	data := BuildStruct{
		PageTitle: "(GAME)",
	}

	data.Version = updater.Release

	if os.Args[len(os.Args)-1] == "-block" {
		data.Insert = false
	} else {
		data.Insert = true
	}

	if r.Method == "GET" {
		tmpl.Execute(w, data)
		log.Println("Test 1")
	} else {
		log.Println("Test 2")
		r.ParseForm()
		var commands = r.Form["commands"][0]
		util.WriteLogToFile(util.WebLog, "SEARCH: "+commands)
	}
}

func answerFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	r.ParseForm()
	action := r.FormValue("action")
	query := r.FormValue("query")
	aiResponse := r.FormValue("airesponse")

	tmpl := template.Must(template.ParseFS(tplFolder, "templates/home.html"))
	data := BuildStruct{
		PageTitle: "(SCMD)",
		Version:   updater.Release,
	}
	if os.Args[len(os.Args)-1] == "-block" {
		data.Insert = false
	} else {
		data.Insert = true
	}

	switch action {
	case "save":
		if strings.HasPrefix(strings.TrimSpace(aiResponse), "## ID:") {
			data.SaveStatus = "already"
		} else {
			success, err := database.AddCommand(query, aiResponse, ai.GetBestEmbedding)
			if err != nil || !success {
				log.Printf("Error saving AI response: %v", err)
				data.SaveStatus = "error"
			} else {
				data.SaveStatus = "saved"
			}
		}
		tokenStr := "Usage Not Tracked"
		aiPage := fmt.Sprintf("## AI-Generated Response\n\n**TOKEN:** %s\n\n%s", tokenStr, aiResponse)
		data.Pages = []string{aiPage}
		data.PageQuery = query
		tmpl.Execute(w, data)

	case "retry":
		results, newResponse, aiTokens, err := ai.SmartSearch(query, true)
		_ = results
		if err != nil || newResponse == "" {
			data.Pattern = "Could not find a better answer"
			tmpl.Execute(w, data)
			return
		}
		tokenStr := "Usage Not Tracked"
		if aiTokens > 0 {
			tokenStr = strconv.Itoa(aiTokens)
		}
		aiPage := fmt.Sprintf("## AI-Generated Response\n\n**TOKEN:** %s\n\n%s", tokenStr, newResponse)
		data.Pages = []string{aiPage}
		data.PageQuery = query
		tmpl.Execute(w, data)

	default:
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session_id"); err == nil {
		if _, exists := sessionStore.GetSession(cookie.Value); exists {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl := template.Must(template.ParseFS(tplFolder, "templates/login.html"))

	data := BuildStruct{
		PageTitle: "Login",
		Version:   updater.Release,
	}

	if r.Method == "GET" {
		tmpl.Execute(w, data)
	} else {
		r.ParseForm()
		email := r.FormValue("email")
		apiKey := r.FormValue("api_key")

		authenticated, err := authenticateUser(email, apiKey)
		if err != nil {
			log.Printf("Authentication error: %v", err)
			data.Status = "error"
			tmpl.Execute(w, data)
			return
		}

		if !authenticated {
			log.Printf("Failed login attempt for email: %s", email)
			data.Status = "failed"
			tmpl.Execute(w, data)
			return
		}

		sessionID, err := sessionStore.CreateSession(email)
		if err != nil {
			log.Printf("Session creation error: %v", err)
			data.Status = "error"
			tmpl.Execute(w, data)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Path:     "/",
			MaxAge:   86400,
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})

		log.Printf("Successful login for email: %s", email)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func logoutPage(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		sessionStore.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
