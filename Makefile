BINARY_NAME=merkle-generator
PACKAGE=merkle-generator

.PHONY: build clean test run help install deps

# Build the binary
build:
	go build -o $(BINARY_NAME) .

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)

# Run tests
test:
	go test ./... -v

# Run tests with coverage
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Install dependencies
deps:
	go mod tidy
	go mod download

# Install the binary to GOPATH/bin
install: build
	go install .

# Run the application (example usage)
run: build
	./$(BINARY_NAME) --help

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Run example commands
example: build
	@echo "=== Generating root for alice, bob, charlie ==="
	./$(BINARY_NAME) root alice bob charlie
	@echo ""
	@echo "=== Generating proof for alice ==="
	./$(BINARY_NAME) proof alice alice bob charlie
	@echo ""
	@echo "=== Hashing data ==="
	./$(BINARY_NAME) hash alice

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  deps          - Install/update dependencies"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  run           - Run the application (shows help)"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  example       - Run example commands"
	@echo "  help          - Show this help"