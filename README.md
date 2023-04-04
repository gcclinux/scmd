# scmd (Search Command)

Simple search command App that gives the possibility to find commands or store commands locally, this app will evolve and have a web interface in the near future.<BR>

Release: 1.0.0 - (18-02-2023) Initial SCMD CLI & Web UI<BR>
Release: 1.0.1 - (19-02-2023) Recompiled with updated tardigrade-mod v0.2.0<BR>
Release: 1.0.2 - (26-02-2023) Minor cosmetic changes in the search UI<BR>
Release: 1.1.0 - (05-03-2023) Added binary upgrade option in the menu!<BR>
Release: 1.2.0 - (12-03-2023) Added option to specific what port to open the Web UI<BR>
Release: 1.3.0 - (19-03-2023) Added option to save or display functions also<BR>
Release: 1.3.1 - (26-03-2023) Check if command already exist + cosmetics<BR>
Release: 1.3.2 - (01-04-2023) Created the Help page and added search login.<BR>
Release: 1.3.3 - (05-04-2023) Minor cosmetics on help page (before annual leave).<BR>

> Display this help menu
```
Usage: 	 scmd-Linux-x86_64 --help
```
> Opens the web UI with default Port: "3333" 
```
Usage: 	 scmd-Linux-x86_64 --web
```
> Opens the web UI with alternative Port: (-service won't launch the browser)
```
Usage: 	 scmd-Linux-x86_64 --web -port [port] -service
```
> Show local and available scmd version
```
Usage: 	 scmd-Linux-x86_64 --version
```
> Create a copy for the commands database and save it in Home folder
```
Usage: 	 scmd-Linux-x86_64 --copydb
```
> Download all available commands database from online (override locally saved commands)
```
Usage: 	 scmd-Linux-x86_64 --download
```
> Download and upgrade the latest version of the scmd application binary
```
Usage: 	 scmd-Linux-x86_64 --upgrade
```
> Search command based on comma separated pattern(s)
```
Usage: 	 scmd-Linux-x86_64 --search "patterns"
```
> Save new command with description in the local database
```
Usage: 	 scmd-Linux-x86_64 --save "command" "description"
```

This app is also enriched by utilising the "tardigrade-mod" database available for download from github.com

Build and compile scmd from source code will require
>go get [github.com/gcclinux/tardigrade-mod](https://github.com/gcclinux/tardigrade-mod)


\* Download and Install - https://go.dev/dl/ <BR>
\* Download and Install - https://git-scm.com/downloads

```
$ git clone https://github.com/gcclinux/scmd.git
$ cd scmd/
$ go get github.com/gcclinux/tardigrade-mod
$ go build -o scmd-$(uname)-$(uname -m) *.go
$ scmd-$(uname)-$(uname -m) --help
```
