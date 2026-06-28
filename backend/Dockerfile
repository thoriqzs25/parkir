FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api

FROM alpine:3.19

RUN apk add --no-cache ca-certificates postgresql-client

RUN mkdir -p /var/backups/parkir

WORKDIR /app

COPY --from=builder /app/api .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./api"]