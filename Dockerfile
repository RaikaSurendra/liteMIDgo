# Build stage
FROM golang:1.24-alpine AS builder

# Install git and other build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
COPY agent/go.mod agent/go.sum ./agent/

# Download dependencies
RUN go mod download
RUN cd agent && go mod download

# Copy source code
COPY . .

# Build the server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o litemidgo .

# Build the agent
RUN cd agent && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o litemidgo-agent .

# Final stage - Server
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S litemidgo && \
    adduser -u 1001 -S litemidgo -G litemidgo

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/litemidgo .

# Copy config files
COPY --from=builder /app/config ./config

# Create directories for logs
RUN mkdir -p /app/logs && chown -R litemidgo:litemidgo /app

# Switch to non-root user
USER litemidgo

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the server
CMD ["./litemidgo", "server-simple"]
