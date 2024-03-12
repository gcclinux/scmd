# Start from the official golang image
FROM golang:1.22 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY download.go ./
COPY helpmenu.go ./
COPY iscode.go ./
COPY LICENSE ./
COPY main.go ./
COPY README.md ./
COPY openurl.go ./
COPY release ./
COPY release.go ./
COPY version.go ./
COPY search.go ./
COPY server.go ./
COPY tools.go ./
COPY upgrade.go ./
COPY savecmd.go ./
COPY templates ./templates

# Build the Go app
RUN CGO_ENABLED=0 go build -o scmd *.go

# Check if scmd is built successfully
RUN if [ ! -f /app/scmd ]; then echo "scmd build failed" && exit 1; fi

RUN touch /app/scmdweb.log && chmod 666 /app/scmdweb.log

# Start a new stage from scratch
FROM debian:buster-slim

# Update package lists and install ca-certificates
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Add Maintainer Info
LABEL maintainer="Ricardo Wagemaker"
LABEL app.name="SCMD"
LABEL app.version="1.3.7"
LABEL app.release="2024-02-25"
LABEL app.build="007"
LABEL app.url="http://localhost:8080"
LABEL app.description="Simple Command Line or Code Search"
LABEL app.license="MIT"
LABEL app.repository="https://github.com/gcclinux/scmd"
LABEL app.dockerhub="gcclinux/scmd:latest"
LABEL app.dockerhub.tag="latest"

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/scmd .
COPY tardigrade.db /app/tardigrade.db

# Expose port 8080 and 8443 to the outside world
EXPOSE 8080 8443

# Command to run the executable
ENTRYPOINT ["./scmd"]