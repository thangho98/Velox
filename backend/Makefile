.PHONY: build run dev clean test lint fmt migrate migrate-status migrate-rollback

APP_NAME := velox
BUILD_DIR := ./bin

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run: build
	$(BUILD_DIR)/$(APP_NAME)

dev:
	go run ./cmd/server

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./... -v -count=1

test-short:
	go test ./... -short

lint:
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, running go vet only"; \
	fi

fmt:
	gofmt -w -s .

migrate:
	go run ./cmd/server migrate up

migrate-status:
	go run ./cmd/server migrate status

migrate-rollback:
	go run ./cmd/server migrate rollback
