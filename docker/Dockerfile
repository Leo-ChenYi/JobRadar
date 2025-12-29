# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with static linking
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o jobradar ./cmd/jobradar

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install CA certificates for HTTPS requests and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from build stage
COPY --from=builder /app/jobradar /app/jobradar

# Create data directory
RUN mkdir -p /app/data

# Set timezone
ENV TZ=Asia/Shanghai

# Create non-root user
RUN adduser -D -u 1000 jobradar
RUN chown -R jobradar:jobradar /app
USER jobradar

# Entry point
ENTRYPOINT ["/app/jobradar"]
CMD ["run", "--config", "/app/config.yaml"]

