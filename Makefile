# Build the application
all: build test

# build:
# 	@echo "Building..."
# 	@go build -o main cmd/api/main.go

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Test the application
test-full:
	@echo "Testing..."
	@go test ./... -v -tags=integration

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

.PHONY: all build run test clean watch
