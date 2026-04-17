# ---------- Stage 1: Build ----------
FROM golang:1.25.3-alpine AS indexer_builder

# Set the working directory
WORKDIR /app

# Ensure a portable, static-ish binary
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application (strip debug info for smaller size)
RUN go build -trimpath -ldflags="-s -w" -o indexer ./cmd/indexer/

# ---------- Stage 2: Final ----------
FROM alpine:3.23

# Set the working directory
WORKDIR /app

# Install runtime dependencies you actually need
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -S indexer_user \
 && adduser -S -G indexer_user -H -s /sbin/nologin indexer_user

# Copy the binary and set ownership
COPY --from=indexer_builder --chown=indexer_user:indexer_user /app/indexer /app/indexer

# Run as non-root user
USER indexer_user

# Set the entrypoint command
ENTRYPOINT ["/app/indexer"]