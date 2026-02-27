# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies for bluetooth
RUN apk add --no-cache build-base bluez-dev

# Copy go mod and sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o main cmd/main.go

# Run stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata bluez dbus

# Copy binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/.env .

# Expose port
EXPOSE 3003

# Run the application
CMD ["./main"]