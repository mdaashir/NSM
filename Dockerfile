# Build stage
FROM golang:1.24.2-alpine AS builder

# Set build arguments
ARG VERSION="dev"
ARG COMMIT="unknown"
ARG BUILD_DATE="unknown"

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

# Build the application with version information
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X github.com/mdaashir/NSM/cmd.Version=${VERSION} -X github.com/mdaashir/NSM/cmd.Commit=${COMMIT} -X github.com/mdaashir/NSM/cmd.BuildDate=${BUILD_DATE}" -o nsm

# Generate shell completions
RUN mkdir -p /app/completions &&
    ./nsm completion bash >/app/completions/nsm.bash &&
    ./nsm completion zsh >/app/completions/nsm.zsh &&
    ./nsm completion fish >/app/completions/nsm.fish

# Security scan stage
FROM golang:1.24.2-alpine AS security-checker
WORKDIR /app
COPY --from=builder /app/go.mod /app/go.sum ./
RUN apk add --no-cache git &&
    go install golang.org/x/vuln/cmd/govulncheck@latest &&
    govulncheck ./...

# Final stage
FROM alpine:latest

# Build-time metadata as defined at https://github.com/opencontainers/image-spec/blob/master/annotations.md
LABEL org.opencontainers.image.title="NSM" \
    org.opencontainers.image.description="Nix Shell Manager - A tool to manage Nix development environments" \
    org.opencontainers.image.url="https://github.com/mdaashir/NSM" \
    org.opencontainers.image.source="https://github.com/mdaashir/NSM" \
    org.opencontainers.image.version="${VERSION}" \
    org.opencontainers.image.created="${BUILD_DATE}" \
    org.opencontainers.image.revision="${COMMIT}" \
    org.opencontainers.image.licenses="MIT"

# Create non-root user
RUN adduser -D nsm

# Install required packages and set up Nix
RUN apk add --no-cache curl xz sudo shadow bash &&
    mkdir -p /nix /etc/nix &&
    chmod 755 /nix &&
    echo "sandbox = false" >/etc/nix/nix.conf &&
    echo "experimental-features = nix-command flakes" >>/etc/nix/nix.conf &&
    chown -R nsm:nsm /nix /etc/nix

# Copy binary and completions from builder
COPY --from=builder --chown=nsm:nsm /app/nsm /usr/local/bin/nsm
COPY --from=builder --chown=nsm:nsm /app/completions/nsm.bash /etc/bash_completion.d/nsm
COPY --from=builder --chown=nsm:nsm /app/completions/nsm.zsh /usr/share/zsh/site-functions/_nsm
COPY --from=builder --chown=nsm:nsm /app/completions/nsm.fish /usr/share/fish/vendor_completions.d/nsm.fish

# Create config directory
RUN mkdir -p /home/nsm/.config/NSM &&
    chown -R nsm:nsm /home/nsm/.config

# Switch to non-root user
USER nsm

# Set up environment
ENV PATH="/usr/local/bin:${PATH}"
WORKDIR /home/nsm

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 CMD [ "nsm", "doctor", "--json" ]

ENTRYPOINT ["nsm"]
CMD ["--help"]
