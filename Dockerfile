# Build stage
FROM golang:1.25.3-alpine AS builder

RUN apk add --no-cache curl

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code, excluding test files
COPY gen/ ./gen/
COPY src/ ./src/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /geovision ./src/main.go

# Runtime stage
FROM alpine:3.20

# Install CA certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /app

# Copy built binary from builder stage
COPY --from=builder /geovision .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./geovision"]