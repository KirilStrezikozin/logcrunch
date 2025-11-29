build:
	@echo "Building Logcrunch client"
	go build -o bin/logcrunch -v ./cmd/logcrunch

run-server:
	@echo "Running Logcrunch demo server"
	go run ./cmd/server

run-client:
	@echo "Running Logcrunch client"
	go run ./cmd/logcrunch
