templ:
	@echo "Generating templates"
	templ generate

test: templ
	@echo "Running tests"
	go test ./...

lint:
	@echo "Running linter"
	golangci-lint run

build: test templ
	@echo "Building Logcrunch client"
	go build -o bin/logcrunch -v ./cmd/logcrunch

run-client: templ
	@echo "Running Logcrunch client"
	go run ./cmd/logcrunch

run-server:
	@echo "Running Logcrunch demo server"
	go run ./cmd/server
