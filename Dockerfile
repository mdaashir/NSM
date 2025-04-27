# Build stage
FROM golang:1.24.2-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set up workspace
WORKDIR /app

# Copy Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o nsm

# Final stage
FROM alpine:latest

# Create non-root user
RUN adduser -D nsm

# Install required packages and set up Nix
RUN apk add --no-cache curl xz sudo shadow && \
    mkdir -p /nix /etc/nix && \
    chmod 755 /nix && \
    echo "sandbox = false" > /etc/nix/nix.conf && \
    echo "experimental-features = nix-command flakes" >> /etc/nix/nix.conf && \
    chown -R nsm:nsm /nix /etc/nix

# Copy binary from builder
COPY --from=builder --chown=nsm:nsm /app/nsm /usr/local/bin/nsm

# Switch to non-root user
USER nsm

# Set up environment
ENV PATH="/usr/local/bin:${PATH}"
WORKDIR /home/nsm

ENTRYPOINT ["nsm"]
