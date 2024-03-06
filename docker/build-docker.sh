#!/usr/bin/env sh
#
#
echo "### Building docker will rename .git folder temporarely to exclude it from the docker container."
echo "### Press any key to continue or Ctrl+C to stop."
read -n 1 -s -r
echo ""
set -e
mv -fv ../.git ../thisisgit
echo "Copying files to docker folder..."
cp -rv ../download.go ./
cp -rv ../go.mod ./
cp -rv ../go.sum ./
cp -rv ../helpmenu.go ./
cp -rv ../iscode.go ./
cp -rv ../LICENSE ./
cp -rv ../main.go ./
cp -rv ../openurl.go ./
cp -rv ../README.md ./
cp -rv ../release ./
cp -rv ../release.go ./
cp -rv ../savecmd.go ./
cp -rv ../search.go ./
cp -rv ../server.go ./
cp -rv ../tardigrade.db ./
cp -rv ../templates/ ./
cp -rv ../tools.go ./
cp -rv ../upgrade.go ./
cp -rv ../version.go ./
echo ""
echo "Building docker container..."
docker buildx build . -t scmd:latest
mv -fv ../thisisgit ../.git
echo ""
echo "Done building docker container."
echo "To run local container, execute:"
echo "......"
echo "
$ docker run -it --publish 8080:8080 --name SCMD-WEB scmd:latest --web -port 8080 -service
"
echo "......"
echo "
$ docker run -it --publish 8443:8443 --name SCMD-SSL
--volume /cernbot/domain.co.uk/cert.pem:/etc/ssl/certs/cert.pem:ro
--volume /cernbot/domain.co.uk/privkey.pem:/etc/ssl/private/privkey.pem:ro
scmd:latest --ssl -port 8443 -service /etc/ssl/certs/cert.pem /etc/ssl/private/privkey.pem
"
