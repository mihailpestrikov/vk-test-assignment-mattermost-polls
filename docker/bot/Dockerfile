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
    TARANTOOL_PASS=testpass \
    TARANTOOL_SPACE_POLLS=polls \
    TARANTOOL_SPACE_VOTES=votes \
    LOG_LEVEL=info

EXPOSE 8080

CMD ["./pollbot"]