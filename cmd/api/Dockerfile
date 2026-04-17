# ---------- Stage 1: Build ----------
FROM golang:1.25-alpine AS api_builder

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
RUN go build -trimpath -ldflags="-s -w" -o api ./cmd/api/

# ---------- Stage 2: Final ----------
FROM alpine:3.23

# Set the working directory
WORKDIR /app

# Install runtime dependencies you actually need
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -S api_user \
 && adduser -S -G api_user -H -s /sbin/nologin api_user

# Copy the binary and set ownership
COPY --from=api_builder --chown=api_user:api_user /app/api /app/api

# Run as non-root user
USER api_user

# Set the entrypoint command
ENTRYPOINT ["/app/api"]