.PHONY: all build lint test clean

# Binary name
BINARY_NAME=netmon

all: build

build:
	go build -o $(BINARY_NAME) ./cmd/netmon

lint:
	go vet ./...
	go fmt ./...

test:
	go test -v -race ./...

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-linux-*
	rm -f $(BINARY_NAME)-darwin-*
