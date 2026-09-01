.PHONY: init-config run test tidy build clean

init-config:
	@test -f config/config.yaml || cp config/config.example.yaml config/config.yaml
	@test -f .env || cp .env.example .env

run:
	go run ./cmd/server

test:
	go test ./...

tidy:
	go mod tidy

build:
	go build -o bin/server ./cmd/server

clean:
	rm -rf bin
