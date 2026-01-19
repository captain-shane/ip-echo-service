# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy all Go source files
COPY go.mod main.go ./

# Download dependencies (go mod tidy will create go.sum)
RUN go mod tidy && go mod download

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o ip-service .

# Final stage
FROM alpine:latest

WORKDIR /app
RUN apk --no-cache add ca-certificates curl tar && \
    apk upgrade --no-cache

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy binary, startup script, and static assets
COPY --from=builder /app/ip-service .
COPY startup.sh ./
COPY static ./static

# Create geoip directory (databases downloaded at runtime)
RUN mkdir -p ./geoip

# Set permissions
RUN chmod +x ./startup.sh && \
    chown -R appuser:appgroup /app

USER appuser

# Cloud Run uses PORT env var (default 8080)
EXPOSE 8080

# Use startup script to download GeoIP DBs and start service
CMD ["./startup.sh"]
