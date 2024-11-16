.PHONY: build run deps test clean

BINARY_NAME=kraken-portfolio
BUILD_DIR=bin

build:
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/main.go

run:
	@go run ./cmd/main.go

deps:
	@go get github.com/gorilla/websocket
	@go get golang.org/x/term
	@go get github.com/joho/godotenv
	@go mod tidy

test:
	@go test -v ./...

clean:
	@go clean
	@rm -rf $(BUILD_DIR)