build := $(shell git rev-parse --short HEAD)

ld_flags := -ldflags "-X=main.Build=$(build)"

binary := build/addon
os ?= linux
arch ?= amd64

.PHONY: build run clean

build: clean
	@echo "Building..."
	mkdir -p build/
	GOOS=$(os) GOARCH=$(arch) go build $(ld_flags) -o $(binary) main.go
	cp -r web/ build/

run:
	@echo "Running..."
	go run ${ld_flags} main.go

clean:
	@rm -rf build/
