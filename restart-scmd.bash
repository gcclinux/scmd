#!/usr/bin/env bash
#
echo "Stoping the SCMD screen Services"
/usr/bin/screen -S SCMD-SHARED -X quit
sleep 1
/usr/bin/screen -S SCMD-EDIT -X quit
sleep 1
echo "Starting the SCMD services"
cd /server/scmd/ && /usr/bin/screen -dmS SCMD-SHARED /server/scmd/scmd-Linux-aarch64 --ssl -port 3333 -service /server/cernbot/wagemaker.no-ip.co.uk/cert.pem /server/cernbot/wagemaker.no-ip.co.uk/privkey.pem -block
cd /server/scmd/ && /usr/bin/screen -dmS SCMD-EDIT /server/scmd/scmd-Linux-aarch64 --web -port 4444 -service
sleep 1

screen -list | grep SCMD
