FROM golang:1.23

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
