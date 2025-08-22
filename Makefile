.PHONY: help test coverage lint fmt clean install example docs

# Default target
help:
	@echo "GoTrust - Authentication Library for Go"
	@echo ""
	@echo "Available commands:"
	@echo "  make install    Install dependencies"
	@echo "  make test       Run tests"
	@echo "  make coverage   Run tests with coverage"
	@echo "  make lint       Run linters"
	@echo "  make fmt        Format code"
	@echo "  make example    Run basic example"
	@echo "  make clean      Clean build artifacts"
	@echo "  make docs       Generate documentation"

# Install dependencies
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test -v -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linters
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./...; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

# Run basic example
example:
	@echo "Starting basic example server..."
	@if [ ! -f examples/basic/.env ]; then \
		echo "Creating .env file from example..."; \
		cp examples/basic/.env.example examples/basic/.env; \
		echo "Please edit examples/basic/.env with your configuration"; \
	fi
	cd examples/basic && go run main.go

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f coverage.out coverage.html
	rm -rf dist/ build/ bin/
	find . -name "*.test" -delete
	find . -name "*.out" -delete

# Generate documentation
docs:
	@echo "Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "Starting godoc server on http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "godoc not installed. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# Run security audit
audit:
	@echo "Running security audit..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi
	go list -m all | nancy sleuth

# Check for updates
updates:
	@echo "Checking for dependency updates..."
	go list -u -m all

# Build for multiple platforms
build:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build -o dist/gotrust_linux_amd64
	GOOS=darwin GOARCH=amd64 go build -o dist/gotrust_darwin_amd64
	GOOS=windows GOARCH=amd64 go build -o dist/gotrust_windows_amd64.exe
	@echo "Builds complete in dist/"