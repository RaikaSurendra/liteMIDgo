# Build stage
FROM golang:1.24-alpine AS builder

# Install git and other build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the agent
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o litemidgo-agent .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S litemidgo && \
    adduser -u 1001 -S litemidgo -G litemidgo

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/litemidgo-agent .

# Create directories for logs
RUN mkdir -p /app/logs && chown -R litemidgo:litemidgo /app

# Switch to non-root user
USER litemidgo

# Run the agent
CMD ["./litemidgo-agent", "daemon", "--interval", "10"]
