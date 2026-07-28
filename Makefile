.PHONY: all build lint test clean release db-inspect

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

setup:
	@./scripts/setup.sh

db-inspect:
	@./scripts/db-inspect.sh "$(DB_PATH)" "$(SCAN_ID)"

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-linux-*
	rm -f $(BINARY_NAME)-darwin-*


