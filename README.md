# scmd (Search Command)

Simple search command App that gives the possibility to find commands or store commands locally, this app will evolve and have a web interface in the near future.

Release: 1.0.0 - Initial SCMD CLI & Web UI <BR>
Release: 1.0.1 - Recompiled with updated tardigrade-mod v0.2.0 <BR>
Release: 1.0.2 - Minor cosmetic changes in the search UI <BR>
Release: 1.1.0 - Added binary upgrade option in the menu! <BR>
Release: 1.2.0 - Added option to specific what port to open the Web UI <BR>s

> Display this help menu
```
Usage: 	 scmd-Linux-x86_64 --help
```
> Opens the web UI with default Port: "3333" 
```
Usage: 	 scmd-Linux-x86_64 --web
```
> Opens the web UI with alternative Port:
```
Usage: 	 scmd-Linux-x86_64 --web -port [port]
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

Build and compile scmd from source code
>go get [github.com/gcclinux/tardigrade-mod](https://github.com/gcclinux/tardigrade-mod)

```
>> Download and Install - https://go.dev/dl/ 
$ git clone https://github.com/gcclinux/scmd.git
$ go get github.com/gcclinux/tardigrade-mod
$ go build -o bin/scmd *.go
$ bin/scmd --help
```
