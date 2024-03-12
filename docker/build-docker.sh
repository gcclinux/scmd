#!/usr/bin/env sh
#
#
echo ""
echo "Building docker container..."
docker buildx build . -t gcclinux/scmd:latest
echo ""
echo "Done building docker container."
echo "To run local container, execute:"
echo "......"
echo "
$ docker run -it --publish 8080:8080 --name SCMD-WEB gcclinux/scmd:latest --web -port 8080 -service
"
echo "......"
echo "
$ docker run -it --publish 8443:8443 --name SCMD-SSL
--volume /cernbot/domain.co.uk/cert.pem:/etc/ssl/certs/cert.pem:ro
--volume /cernbot/domain.co.uk/privkey.pem:/etc/ssl/private/privkey.pem:ro
scmd:latest --ssl -port 8443 -service /etc/ssl/certs/cert.pem /etc/ssl/private/privkey.pem
"
