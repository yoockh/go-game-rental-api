# ===== Stage 1: Build =====
# Use Go official image to build the application binary
FROM golang:1.24.0 AS builder

# Set working directory inside container
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Change directory to the app folder where main.go is located
WORKDIR /app/app/echo-server

# Build Go binary
RUN go build -o echo-server main.go

# ===== Stage 2: Runtime =====
# Use a lightweight Debian image for running the app
FROM debian:bookworm-slim

# Install ca-certificates for HTTPS support
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy environment file if needed
COPY .env .

# Copy the binary built from the previous stage
COPY --from=builder /app/app/echo-server/echo-server /echo-server

# Expose the port that your application will listen on
EXPOSE 9007

# Command to run the binary when container starts
CMD ["/echo-server"]
