FROM golang:1.20.3-alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

ADD go.mod .
ADD go.sum .

RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o /bank cmd/main.go

FROM alpine:latest

COPY --from=builder /usr/share/zoneinfo/Europe/Moscow /usr/share/zoneinfo/Europe/Moscow
ENV TZ Europe/Moscow

WORKDIR /app

COPY --from=builder /bank /app/bank
COPY --from=builder /build/migration /app/migration
COPY --from=builder /build/docs /app/docs

CMD ["./bank"]