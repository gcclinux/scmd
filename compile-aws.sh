#!/usr/bin/env sh
#
go build -o scmd-$(uname)-$(uname -m) *.go
echo "Setting permission scmd-$(uname)-$(uname -m)"
sudo setcap CAP_NET_BIND_SERVICE=+eip /home/ubuntu/scmd/scmd-$(uname)-$(uname -m)
#
echo "Stoping the SCMD screen services"
/usr/bin/screen -S SCMD80 -X quit
/usr/bin/screen -S SCMD443 -X quit
sleep 1
echo "Starting the SCMD re-compiled services"
/usr/bin/screen -dmS SCMD443 /home/ubuntu/scmd/scmd-Linux-x86_64 --ssl -port 443 -service /home/ubuntu/scmd/crts/cert.pem  /home/ubuntu/scmd/crts/privkey.pem
/usr/bin/screen -dmS SCMD80 /home/ubuntu/scmd/scmd-Linux-x86_64 --web -port 80 -service
sleep 1
#
/usr/bin/screen --list | grep SCMD


