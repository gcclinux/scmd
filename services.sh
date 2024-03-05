#!/usr/bin/env sh

DIR=$(dirname "$0")
echo ""

case $1 in
    start)
        echo "Starting the LOCAL SCMD screen Services" 
        cd $DIR && /usr/bin/screen -dmS SCMD-SHARED $DIR/scmd-Linux-aarch64 --ssl -port 3333 -service /server/cernbot/wagemaker.no-ip.co.uk/cert.pem /server/cernbot/wagemaker.no-ip.co.uk/privkey.pem -block
        cd $DIR && /usr/bin/screen -dmS SCMD-EDIT $DIR/scmd-Linux-aarch64 --web -port 4444 -service
        sleep 1
        ;;
    stop)
        echo "Stoping the LOCAL SCMD screen Services"
        /usr/bin/screen -S SCMD-SHARED -X quit
        sleep 1
        /usr/bin/screen -S SCMD-EDIT -X quit
        sleep 1
        ;;
    restart)
        echo "Stoping the LOCAL SCMD screen Services"
        /usr/bin/screen -S SCMD-SHARED -X quit
        sleep 1
        /usr/bin/screen -S SCMD-EDIT -X quit
        sleep 1
        echo "Starting the LOCAL SCMD services"
        cd $DIR && /usr/bin/screen -dmS SCMD-SHARED $DIR/scmd-Linux-aarch64 --ssl -port 3333 -service /server/cernbot/wagemaker.no-ip.co.uk/cert.pem /server/cernbot/wagemaker.no-ip.co.uk/privkey.pem -block
        cd $DIR && /usr/bin/screen -dmS SCMD-EDIT $DIR/scmd-Linux-aarch64 --web -port 4444 -service
        sleep 1
        ;;
    *)
        echo "Usage: $0 {start|stop|restart}"
	    echo ""
        exit 1
esac


echo ""
echo "Running LOCAL SCMD screen Services"
echo ""
/usr/bin/screen -list | grep SCMD
echo ""
