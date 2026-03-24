package server

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gcclinux/scmd/internal/ai"
	"github.com/gcclinux/scmd/internal/database"
	"github.com/gcclinux/scmd/internal/util"
)

// BuildStruct holds template data for web pages.
type BuildStruct struct {
	PageTitle       string
	Pattern         string
	Id              int
	Key             string
	Data            string
	CmdTitle        string
	DescTitle       string
	Version         string
	Return          string
	Status          string
	CmdFunc         string
	AllData         []string
	Code            []string
	Insert          bool
	AIResponse      string
	Pages           []string
	PageQuery       string
	SaveStatus      string
	AIProviderLabel string
}

var tplFolder embed.FS

// SetTemplates sets the embedded filesystem for templates.
func SetTemplates(fs embed.FS) {
	tplFolder = fs
}

// Routes starts the web server with all HTTP/HTTPS configuration.
func Routes() {
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDB()

	ai.InitProviders()

	wg := new(sync.WaitGroup)
	wg.Add(2)

	HTTP := 3333
	browser := true
	SSL := true
	CRT := ""
	KEY := ""

	count := len(os.Args)

	if count == 2 {
		if os.Args[1] == "--web" {
			SSL = false
		} else {
			wrongSyntax()
			os.Exit(1)
		}
	}

	if count == 3 && os.Args[count-1] != "-block" {
		wrongSyntax()
		os.Exit(1)
	} else if count == 3 && os.Args[count-1] == "-block" {
		if os.Args[1] == "--web" {
			SSL = false
		} else {
			wrongSyntax()
			os.Exit(1)
		}
	}

	if count == 4 && os.Args[count-1] != "-block" {
		if os.Args[2] == "-port" && util.IsInt(os.Args[3]) {
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
		if os.Args[1] == "--web" {
			SSL = false
		}
		if os.Args[2] == "-port" && util.IsInt(os.Args[3]) {
			HTTP, _ = strconv.Atoi(os.Args[3])
		} else if os.Args[2] == "-service" {
			browser = false
		}
		if os.Args[4] == "-service" {
			browser = false
			if SSL {
				wrongSyntax()
				os.Exit(1)
			}
		}
		if SSL {
			if os.Args[2] == "-service" {
				CRT = os.Args[3]
				KEY = os.Args[4]
			} else {
				CRT = os.Args[2]
				KEY = os.Args[3]
			}
		}
	}

	if count == 6 && os.Args[count-1] == "-block" {
		if os.Args[1] == "--web" {
			SSL = false
		}
		if os.Args[2] == "-port" && util.IsInt(os.Args[3]) {
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

	if count == 6 && os.Args[count-1] != "-block" {
		if os.Args[1] == "--web" {
			SSL = false
		}
		if os.Args[2] == "-port" && util.IsInt(os.Args[3]) {
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
	} else if count == 6 && os.Args[count-1] == "-block" {
		if os.Args[2] == "-service" {
			browser = false
		}
		if !SSL {
			wrongSyntax()
			os.Exit(1)
		} else {
			CRT = os.Args[3]
			KEY = os.Args[4]
		}
	}

	if count == 7 && os.Args[count-1] != "-block" {
		if os.Args[2] == "-port" && util.IsInt(os.Args[3]) {
			HTTP, _ = strconv.Atoi(os.Args[3])
		} else {
			wrongSyntax()
			os.Exit(1)
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
	} else if count == 7 && os.Args[count-1] == "-block" {
		if os.Args[2] == "-port" && util.IsInt(os.Args[3]) {
			HTTP, _ = strconv.Atoi(os.Args[3])
		} else {
			wrongSyntax()
			os.Exit(1)
		}
		if SSL {
			CRT = os.Args[4]
			KEY = os.Args[5]
		} else {
			wrongSyntax()
			os.Exit(1)
		}
	}

	if count == 8 && os.Args[count-1] == "-block" {
		if os.Args[2] == "-port" && util.IsInt(os.Args[3]) {
			HTTP, _ = strconv.Atoi(os.Args[3])
		} else {
			wrongSyntax()
			os.Exit(1)
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

	// Register static file handler
	fs := http.FS(tplFolder)
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(fs)))

	// Public routes
	http.HandleFunc("/", homePage)
	if os.Args[count-1] != "-block" {
		http.HandleFunc("/add", addPage)
	}
	http.HandleFunc("/game", gamePage)
	http.HandleFunc("/help", helpPage)
	http.HandleFunc("/answer-feedback", answerFeedback)

	if browser {
		if SSL {
			log.Println("Starting scmd web HTTPS UI on port", HTTP)
			util.OpenBrowser(fmt.Sprintf("https://%s:%v", util.GetOutboundIP(), HTTP))
			err := http.ListenAndServeTLS(fmt.Sprintf(":%v", HTTP), CRT, KEY, nil)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Println("Starting scmd web HTTP UI on port", HTTP)
			util.OpenBrowser(fmt.Sprintf("http://%s:%v", util.GetOutboundIP(), HTTP))
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
			wg.Done()
		}()
	}

	wg.Wait()
}

func wrongSyntax() {
	fmt.Println()
	fmt.Println("Usage: \t", getName(), "--help")
	fmt.Println()
}

func getName() string {
	return os.Args[0]
}
