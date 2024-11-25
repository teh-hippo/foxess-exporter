# Build
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o foxess-exporter

# Tiny Container
FROM scratch
COPY --from=builder /app/foxess-exporter /foxess-exporter
EXPOSE 6787
CMD ["/foxess-exporter"]
