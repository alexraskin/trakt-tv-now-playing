FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

RUN adduser -D -u 1000 appuser

COPY --from=builder /app/main .

RUN chown appuser:appuser /app/main

USER appuser

EXPOSE 8080

CMD ["./main"] 