.PHONY: build run test lint

build:
	go build -o ledger ./cmd/app

run:
	go run ./cmd/app

test:
	go test ./...

lint:
	go vet ./...
