# Stage 1: Build
FROM golang:1.21-alpine AS builder

# Install build dependencies for go-sqlite3
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy and initialize module
COPY main.go .
RUN go mod init liteform && \
    go get github.com/mattn/go-sqlite3

# Build with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o liteform .

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite-libs tzdata

# Set timezone
ENV TZ=Asia/Shanghai

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/liteform .

# Create data directory for SQLite persistence
RUN mkdir data

# Expose default port
EXPOSE 8255

# Environment variables (Defaults)
ENV ADMIN_USER=admin
ENV ADMIN_PASS=""
ENV PORT=8255

CMD ["./liteform"]
