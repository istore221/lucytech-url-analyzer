FROM golang:1.24-alpine AS builder

RUN apk update && apk add --no-cache ca-certificates tzdata make

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download


COPY . .

RUN make quality

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/bin/app ./cmd/cli/main.go



FROM scratch


COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo


COPY --from=builder /app/bin/app /app

ENTRYPOINT ["/app"]