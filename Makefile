.PHONY: test lint coverage benchmark clean build all

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

# Install development dependencies
dev-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run all tests and generate coverage report
test-all: lint test benchmark coverage
