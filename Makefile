.PHONY: build

test:
	go test -race -timeout 10s ./...

build:
	CGO_ENABLED=0 go build -ldflags "-s -w" -o ./out/reddit-bot ./cmd/reddit-bot

all: test build

run: build
	./out/reddit-bot -c .dev/reddit-bot/config.yml
