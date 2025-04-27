.PHONY: test lint coverage benchmark clean build all version-bump build-parallel build-all test-cached

# Use available CPU cores for parallel builds
NUMPROC := $(shell nproc || echo 4)

# Default target
all: lint test build

# Run all tests
test:
	go test -v ./tests/unit/...
	go test -v ./tests/integration/...

# Run linter
lint:
	golangci-lint run

# Run benchmarks
benchmark:
	go test -v -bench=. ./tests/benchmark/...

# Generate test coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f coverage.out coverage.html
	rm -f Output/NSM.exe
	go clean -testcache

# Build the binary
build:
	go build -o Output/NSM.exe

# Build with parallelization
build-parallel:
	go build -p $(NUMPROC) -o Output/NSM.exe

# Build for all platforms in parallel
build-all:
	GOOS=linux GOARCH=amd64 go build -o Output/nsm-linux-amd64 &
	GOOS=linux GOARCH=arm64 go build -o Output/nsm-linux-arm64 &
	GOOS=darwin GOARCH=amd64 go build -o Output/nsm-darwin-amd64 &
	GOOS=darwin GOARCH=arm64 go build -o Output/nsm-darwin-arm64 &
	GOOS=windows GOARCH=amd64 go build -o Output/nsm-windows-amd64.exe &
	wait

# Run tests with caching
test-cached:
	go test -count=1 -v ./tests/unit/... -test.cache
	go test -count=1 -v ./tests/integration/... -test.cache

# Install development dependencies
dev-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run all tests and generate coverage report
test-all: lint test benchmark coverage

# Bump version across all files (usage: make version-bump VERSION=1.2.0)
version-bump:
ifndef VERSION
	$(error VERSION is required. Usage: make version-bump VERSION=1.2.0)
endif
	./scripts/bump-version.sh $(VERSION)
