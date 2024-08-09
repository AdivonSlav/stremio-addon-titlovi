FROM golang:1.22-alpine3.20

WORKDIR /app

# Copy over go.mod and go.sum and install dependencies.
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy over everything.
COPY . .
RUN go build -o ./addon

# Expose the default port.
EXPOSE 5555

# Run
CMD ["./addon"]