# Build stage
FROM golang:1.25.2-alpine AS builder

# Install build dependencies (for CGO and sqlite3)
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o bin/schools-be cmd/api/main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs tzdata

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/schools-be .

# Create directories for data and cache
RUN mkdir -p /app/data /app/cache

# Expose port
EXPOSE 8080

# Run the application
CMD ["./schools-be"]

