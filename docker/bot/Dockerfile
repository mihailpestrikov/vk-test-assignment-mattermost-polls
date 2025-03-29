FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pollbot ./cmd/pollbot

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/pollbot .

RUN mkdir -p /app/logs

ENV PORT=8080 \
    GIN_MODE=release \
    TARANTOOL_HOST=tarantool \
    TARANTOOL_PORT=3301 \
    TARANTOOL_USER=guest \
    TARANTOOL_SPACE_POLLS=polls \
    TARANTOOL_SPACE_VOTES=votes \
    LOG_LEVEL=info \
    MATTERMOST_URL=http://mattermost:8065 \
    DEFAULT_POLL_DURATION=86400 \
    MAX_OPTIONS=10 \
    HOST=0.0.0.0 \
    REQUEST_TIMEOUT=30 \
    APP_ENV=production

EXPOSE 8080

CMD ["./pollbot"]