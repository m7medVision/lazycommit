BINARY := lazycommit

.PHONY: build test lint install clean

build:
	go build -o $(BINARY) .

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install .

clean:
	rm -f $(BINARY)
	rm -rf dist
