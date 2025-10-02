# Build
FROM docker.io/golang:1.25-alpine AS builder
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o foxess-exporter -trimpath -ldflags="-s -w"

# Tiny Container
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/foxess-exporter /usr/bin/foxess-exporter
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
EXPOSE 2112
ENTRYPOINT [ "/usr/bin/foxess-exporter" ]
CMD [ "serve" ]