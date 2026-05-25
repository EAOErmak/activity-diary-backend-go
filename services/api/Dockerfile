FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /bin/api ./cmd/api

FROM alpine:3.21
WORKDIR /app
COPY --from=build /bin/api /usr/local/bin/api
COPY --from=build /app/internal /app/internal
EXPOSE 18080
CMD ["api"]
