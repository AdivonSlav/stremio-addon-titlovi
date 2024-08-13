# Titlovi.com Unofficial Addon for Stremio
This is an unofficial Stremio addon that returns subtitle results from [Titlovi.com](https://titlovi.com), written in Go.

## Access
Simply open the following URL, configure your Titlovi.com credentials and install the addon.

https://titlovi-unofficial-addon-production.up.railway.app/configure

## Running locally
Ensure you have Go installed. To run locally, it is necessary to just run the following commands:
```bash
go mod download && go mod verify
PORT=5555 go run main.go
```
Alternatively, the repository contains a Dockerfile which can be used to build an image and run the addon in a container.
