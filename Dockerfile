# Build stage
FROM golang:1.23-alpine

WORKDIR /app

RUN apk add --no-cache make

# Copy over everything.
COPY . .
RUN go mod download && go mod verify
RUN make build

# Run stage
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the binary from the build stage
COPY --from=build /app/build/addon .

# Expose the default port.
EXPOSE 5555

# Set the entrypoint command
ENTRYPOINT ["/app/addon"]
