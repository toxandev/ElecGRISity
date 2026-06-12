# Default recipe to run when just is invoked without arguments
set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]

default: build

# Variables
BIN_NAME := "elecgrisity-server"
MAIN_PKG := "cmd/elecgrisity/main.go"

# Build the Go binary (Windows)
[windows]
build:
	@echo "Building {{BIN_NAME}}.exe..."
	go build -o {{BIN_NAME}}.exe {{MAIN_PKG}}
	@echo "Build complete."

# Build the Go binary (Unix)
[unix]
build:
	@echo "Building {{BIN_NAME}}..."
	go build -o {{BIN_NAME}} {{MAIN_PKG}}
	@echo "Build complete."

# Run the application (Windows)
[windows]
run: build
	@echo "Starting {{BIN_NAME}}.exe..."
	.\{{BIN_NAME}}.exe

# Run the application (Unix)
[unix]
run: build
	@echo "Starting {{BIN_NAME}}..."
	./{{BIN_NAME}}

# Run all Go tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean up build artifacts (Windows)
[windows]
clean:
	@echo "Cleaning up..."
	if (Test-Path {{BIN_NAME}}.exe) { Remove-Item -Force {{BIN_NAME}}.exe }
	if (Test-Path {{BIN_NAME}}) { Remove-Item -Force {{BIN_NAME}} }
	go clean

# Clean up build artifacts (Unix)
[unix]
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
