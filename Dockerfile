# Build stage
FROM golang:1.24.2-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set up workspace
WORKDIR /app

# Copy Go module files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies in a separate layer
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Run tests and security checks
RUN go test -v ./tests/unit/... && \
    go vet ./...

# Build the application with security flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s -extldflags=-static" \
    -trimpath -o nsm

# Final stage
FROM alpine:latest AS final

# Update packages and add security patches
RUN apk upgrade --no-cache && \
    apk add --no-cache ca-certificates tzdata curl xz sudo shadow && \
    rm -rf /var/cache/apk/*

# Create non-root user
RUN adduser -D nsm && \
    echo "nsm ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Install required packages and set up Nix
RUN apk add --no-cache curl xz sudo shadow && \
    mkdir -p /nix /etc/nix && \
    chmod 755 /nix && \
    echo "sandbox = false" > /etc/nix/nix.conf && \
    echo "experimental-features = nix-command flakes" >> /etc/nix/nix.conf && \
    chown -R nsm:nsm /nix /etc/nix

# Copy binary from builder with explicit permissions
COPY --from=builder --chown=nsm:nsm /app/nsm /usr/local/bin/nsm
RUN chmod 755 /usr/local/bin/nsm

# Switch to non-root user
USER nsm

# Set up environment
ENV PATH="/usr/local/bin:${PATH}" \
    HOME="/home/nsm" \
    TZ=UTC

WORKDIR /home/nsm

# Health check
HEALTHCHECK --interval=30s --timeout=3s \
    CMD nsm doctor || exit 1

ENTRYPOINT ["nsm"]
