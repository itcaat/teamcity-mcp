# Build stage
FROM golang:1.23-alpine AS builder

# Install certificates and git for private dependencies
RUN apk --no-cache add ca-certificates git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o teamcity-mcp ./cmd/server

# Final stage
FROM scratch

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/teamcity-mcp /teamcity-mcp

# Expose port
EXPOSE 8123

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/teamcity-mcp", "-transport", "http", "&", "curl", "-f", "http://localhost:8123/healthz", "||", "exit", "1"]

# Run the application
ENTRYPOINT ["/teamcity-mcp"] 