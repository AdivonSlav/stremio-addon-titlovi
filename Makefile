BUILD := $(shell git rev-parse --short HEAD)
PROJECTNAME := stremio-unofficial-addon

LDFLAGS=-ldflags "-X=main.Build=$(BUILD)"

build:
	@echo "Building..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o build/addon main.go

run:
	@echo "Running..."
	go run ${LDFLAGS} main.go
