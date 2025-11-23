.PHONY: build run test clean install-deps setup-auth help

# Build the binary (release)
build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -o cc-server ./cmd/server

# Build for current OS (development)
build-local:
	go build -o cc-server ./cmd/server

# Run the server locally
run: build-local
	./cc-server

# Run with custom config
run-with-config:
	go run cmd/server/main.go --config ~/.config/cc/config.json

# Setup authentication (interactive)
setup-auth:
	@echo "Setting up authentication for Command Center v0.3.0"
	@read -p "Enter username: " username; \
	read -s -p "Enter password: " password; \
	echo ""; \
	go run cmd/server/main.go --username $$username --password $$password

# Run tests
test:
	go test ./...

# Run tests with verbose output
test-v:
	go test -v ./...

# Run tests with coverage
test-cover:
	go test ./... -cover

# Clean build artifacts
clean:
	rm -f cc-server
	rm -f cc.db cc.db-shm cc.db-wal
	rm -f command-center-*.tar.gz
	rm -rf ~/.config/cc/backups/

# Install Go dependencies
install-deps:
	go mod download
	go mod tidy

# Create release package
release: build
	tar -czf command-center-v0.3.0.tar.gz \
		cc-server \
		web/ \
		migrations/ \
		examples/ \
		config.example.json \
		README.md \
		CLAUDE.md

# Development - run with auto-reload (requires air)
dev:
	air

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Show help
help:
	@echo "Command Center v0.3.0 - Makefile Targets"
	@echo ""
	@echo "  make build       - Build release binary (linux/amd64)"
	@echo "  make build-local - Build for current OS"
	@echo "  make run         - Build and run server"
	@echo "  make test        - Run all tests"
	@echo "  make test-cover  - Run tests with coverage"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make setup-auth  - Setup authentication"
	@echo "  make release     - Create release tarball"
	@echo "  make help        - Show this help"
