# Build
FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o foxess-exporter

# Tiny Container
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/foxess-exporter /foxess-exporter
EXPOSE 2112
CMD ["/foxess-exporter"]
