FROM golang:1.27-alpine AS builder

ARG APP_PATH=./cmd/cli/main.go

RUN apk update && apk add --no-cache ca-certificates tzdata make

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download


RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.64.5

COPY . .

#RUN make quality

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/bin/app ${APP_PATH}



FROM scratch


COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo


COPY --from=builder /app/bin/app /app

ENTRYPOINT ["/app"]