# Build stage
FROM golang:1.18-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o fai-echo ./cmd/fai-echo

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/fai-echo .
COPY --from=builder /app/.env .
EXPOSE 8000
CMD ["./fai-echo"]