# Default recipe to run when just is invoked without arguments
default: build

# Variables
BIN_NAME := "elecgrisity-server"
MAIN_PKG := "cmd/elecgrisity/main.go"

# Build the Go binary
build:
	@echo "Building {{BIN_NAME}}..."
	go build -o {{BIN_NAME}} {{MAIN_PKG}}
	@echo "Build complete."

# Run the application
run: build
	@echo "Starting {{BIN_NAME}}..."
	./{{BIN_NAME}}

# Run all Go tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	rm -f {{BIN_NAME}} {{BIN_NAME}}.exe
	go clean

# Format Go code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Tidy Go modules
tidy:
	@echo "Tidying go.mod..."
	go mod tidy

# Run all pre-commit checks (format, tidy, test)
check: fmt tidy test
	@echo "All checks passed!"
