FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o devinsync_notifier

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/devinsync_notifier /app/devinsync_notifier

EXPOSE 8080

ENTRYPOINT ["/app/devinsync_notifier"]
