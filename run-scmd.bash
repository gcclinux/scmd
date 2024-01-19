#!/usr/bin/env bash
#
cd /server/scmd/ && /usr/bin/screen -dmS SCMD /server/scmd/scmd-Linux-aarch64 --ssl -port 3333 -service /home/ricardo/scripts/PlexP12/cert.pem /home/ricardo/scripts/PlexP12/privkey.pem -block
cd /server/scmd/ && /usr/bin/screen -dmS SCMDL /server/scmd/scmd-Linux-aarch64 --web -port 4444 -service
