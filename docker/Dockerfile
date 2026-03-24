# Start from the official golang image
FROM golang:1.22 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the entire source tree
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY release ./
COPY LICENSE ./
COPY README.md ./

# Build the Go app
RUN CGO_ENABLED=0 go build -o scmd ./cmd/scmd/

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
LABEL app.version="2.0.5"
LABEL app.description="Simple Command Line or Code Search"
LABEL app.license="MIT"
LABEL app.repository="https://github.com/gcclinux/scmd"

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/scmd .

# Expose port 8080 and 8443 to the outside world
EXPOSE 8080 8443

# Command to run the executable
ENTRYPOINT ["./scmd"]
