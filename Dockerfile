# Build stage
FROM golang:1.25-bookworm AS builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y git gcc libc6-dev

# Copy all go module files
COPY go.mod go.sum ./
# Download dependencies
RUN go mod tidy

# Copy the rest of the source code
COPY . .

# Build the main application
# CGO_ENABLED=1 is REQUIRED for confluent-kafka-go
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/main .

# Runtime stage
FROM debian:bookworm-slim

# Add runtime dependencies
RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /app/

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Copy the static frontend
COPY static ./static

# Expose the port
EXPOSE 8080

# Run the binary
CMD ["./main"]