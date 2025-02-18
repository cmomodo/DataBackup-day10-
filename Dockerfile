# --- Stage 1: Build the application ---
FROM golang:1.23.5-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for reliable caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
# Replace main.go with your actual entry point if needed
RUN go build -o main main.go

# --- Stage 2: Create final container image ---
FROM alpine:latest

# Set working directory in the final image
WORKDIR /app

# Copy compiled binary from builder
COPY --from=builder /app/main /app/

# (Optional) Copy any static files (e.g., HTML templates) if needed
# COPY --from=builder /app/templates ./templates

# Do not copy .env into the container for security reasons. Instead,
# set environment variables at runtime or use a secret manager.

# Expose a port (e.g., 8080 if your app listens on 8080)
EXPOSE 8080

# Run the binary
CMD ["./main"]