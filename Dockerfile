# Build stage
FROM golang:1.22-alpine AS builder

# Install security updates and required packages
RUN apk update && apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# Create non-root user for build
RUN adduser -D -s /bin/sh -u 1001 appuser

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the binary with security flags
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o markdocify \
    ./cmd/markdocify

# Production stage
FROM scratch

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd

# Copy binary
COPY --from=builder /app/markdocify /markdocify

# Use non-root user
USER appuser

# Set entrypoint
ENTRYPOINT ["/markdocify"]

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/markdocify", "--version"]

# Labels for metadata
LABEL org.opencontainers.image.title="markdocify"
LABEL org.opencontainers.image.description="Comprehensively scrape documentation sites into beautiful, LLM-ready Markdown"
LABEL org.opencontainers.image.url="https://github.com/vladkampov/markdocify"
LABEL org.opencontainers.image.source="https://github.com/vladkampov/markdocify"
LABEL org.opencontainers.image.licenses="MIT"