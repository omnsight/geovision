# Build stage
FROM golang:1.25.3-alpine AS builder

RUN apk add --no-cache curl

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
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

# Set environment variables with defaults
ENV GRPC_PORT=9090
ENV SERVER_PORT=8080
ENV ARANGO_URL=http://localhost:8529
ENV ARANGO_DB=geovision
ENV ARANGO_USERNAME=root
ENV ARANGO_PASSWORD=0123

# Run the application
CMD ["./geovision"]