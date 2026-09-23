.PHONY: run test test-race build lint fmt clean docker-build docker-run

# Run the app locally
run:
	go run .

# Run all tests
test:
	go test ./... -v -count=1

# Run tests with race detector
test-race:
	go test ./... -v -race -count=1

# Build the binary
build:
	CGO_ENABLED=0 go build -o bin/task-api .

# Run go vet (static analysis)
lint:
	go vet ./...

# Format source files
fmt:
	go fmt ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build Docker image
docker-build:
	docker build -t task-api:latest .

# Run Docker container
docker-run:
	docker run -p 8080:8080 task-api:latest
