# syntax=docker/dockerfile:1
ARG GO_VERSION=1.25
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

# deps cache
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG APP_ENV=prod
ENV APP_ENV=${APP_ENV}

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/core-service "./cmd/main.go"

FROM alpine:3.20 AS runtime
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /out/core-service ./core-service
COPY --from=builder /app/config ./config

ARG APP_ENV=prod
ENV APP_ENV=${APP_ENV}

EXPOSE 5010 5001 3001 8002

CMD ["./core-service"]
