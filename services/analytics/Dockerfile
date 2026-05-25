FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /bin/analytics ./cmd/analytics

FROM alpine:3.21
WORKDIR /app
COPY --from=build /bin/analytics /usr/local/bin/analytics
EXPOSE 18081
CMD ["analytics"]
