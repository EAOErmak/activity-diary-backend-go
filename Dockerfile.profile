FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o profile-service ./profile-service

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/profile-service ./profile-service

EXPOSE 8081

CMD ["./profile-service"]
