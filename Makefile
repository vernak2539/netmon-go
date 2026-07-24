.PHONY: all build lint test clean release

# Binary name
BINARY_NAME=netmon-go

all: build

build:
	go build -o $(BINARY_NAME) ./cmd/netmon

lint:
	go vet ./...
	go fmt ./...

test:
	go test -v -race ./...

release:
	@./scripts/release.sh $(VERSION)

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-linux-*
	rm -f $(BINARY_NAME)-darwin-*

