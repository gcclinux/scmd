#!/usr/bin/env sh
#
go build -o scmd-$(uname)-$(uname -m) *.go
echo "Setting permission scmd-$(uname)-$(uname -m)"
sudo setcap CAP_NET_BIND_SERVICE=+eip /home/ubuntu/scmd/scmd-$(uname)-$(uname -m)
