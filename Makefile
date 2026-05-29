.PHONY: run build deps

deps:
	go mod tidy

build:
	go build -o bin/ai-assistant ./cmd/

run:
	go run ./cmd/

debug:
	AI_LOG_LEVEL=debug go run ./cmd/
