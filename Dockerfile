# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o nsm main.go

# Final stage
FROM alpine:latest

# Install nix dependencies
RUN apk add --no-cache curl xz

# Install Nix
RUN curl -L https://nixos.org/nix/install | sh

# Copy the binary from builder
COPY --from=builder /app/nsm /usr/local/bin/nsm

# Create default configuration directory
RUN mkdir -p /root/.config/nsm

# Set the entrypoint
ENTRYPOINT ["nsm"]
CMD ["--help"]
