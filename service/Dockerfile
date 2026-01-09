# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o ip-service .

# Final stage
FROM alpine:latest

WORKDIR /app
RUN apk --no-cache add ca-certificates wget && \
    apk upgrade --no-cache

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy binary and static assets
COPY --from=builder /app/ip-service .
COPY static ./static
# Copy GeoIP if present
COPY geoip ./geoip 

# Set permissions
RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080 8443

# Default command
CMD ["./ip-service", "-addr", ":8080", "-static", "./static"]
