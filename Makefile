# Makefile for lazycommit

BINARY=lazycommit
VERSION=0.1.0

# Build the application
build:
	go build -o ${BINARY} main.go

# Install the application
install:
	go install

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f ${BINARY}

# Build for multiple platforms
build-all: build-linux build-mac build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build -o ${BINARY}-linux-amd64 main.go

build-mac:
	GOOS=darwin GOARCH=amd64 go build -o ${BINARY}-darwin-amd64 main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o ${BINARY}-windows-amd64.exe main.go

# Format source code
fmt:
	go fmt ./...

# Run vet
vet:
	go vet ./...

.PHONY: build install test clean build-all build-linux build-mac build-windows fmt vet