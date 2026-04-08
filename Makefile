# CoffeeGraph — Build & Development Targets
# Usage: make [target]
#   make build     - Build the binary
#   make test      - Run all tests with race detector
#   make lint      - Run golangci-lint
#   make coverage  - Generate HTML coverage report
#   make all       - Run vet + lint + test + build
#   make snapshot  - Create a local GoReleaser snapshot

.PHONY: build test lint vet clean install all coverage snapshot doctor

BINARY  := coffeegraph
PKG     := ./cmd/coffeegraph
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(PKG)

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run ./...

vet:
	go vet ./...

coverage:
	@mkdir -p .coverage
	go test -race -coverprofile=.coverage/cover.out ./...
	go tool cover -html=.coverage/cover.out -o .coverage/index.html
	@echo "Coverage report: .coverage/index.html"

snapshot:
	goreleaser release --snapshot --clean

doctor: build
	./$(BINARY) init .doctor-test 2>/dev/null || true
	cd .doctor-test && ../$(BINARY) doctor || true
	@rm -rf .doctor-test

clean:
	rm -f $(BINARY)
	rm -rf dist/ .coverage/ .doctor-test/

install: build
	mv $(BINARY) $(GOPATH)/bin/ 2>/dev/null || mv $(BINARY) ~/go/bin/

all: vet lint test build
