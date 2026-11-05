VERSION ?= 0.1.0
BINARY  := dist/alat-gudang.exe

.PHONY: all run test css build-win dist clean

all: css test build-win

css:
	npx tailwindcss -i ./internal/web/static/input.css -o ./internal/web/static/app.css --minify

test:
	go test ./...

run:
	go run ./cmd/server -dev -seed

build-win:
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w -X main.Version=$(VERSION)" -o $(BINARY) ./cmd/server

dist: css test build-win
	mkdir -p dist/data
	cp README.md dist/README.txt

clean:
	rm -rf dist

