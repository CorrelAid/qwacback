# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o qwacback main.go

# Run stage
FROM alpine:3.21
RUN apk add --no-cache ca-certificates && \
    adduser -D -u 1001 appuser
WORKDIR /app
COPY --from=builder /app/qwacback /app/qwacback
COPY --from=builder /app/seed_data /app/seed_data
RUN chown -R appuser:appuser /app

USER appuser
EXPOSE 8080
CMD ["./qwacback", "serve", "--http=0.0.0.0:8080"]
