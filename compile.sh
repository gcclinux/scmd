#!/usr/bin/env sh
#
go build -o scmd-$(uname)-$(uname -m) *.go
echo "scmd-$(uname)-$(uname -m)"
