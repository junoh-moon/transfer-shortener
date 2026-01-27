FROM golang:1.24-alpine AS builder

ARG COMMIT=unknown
ARG BUILD_TIME=unknown

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
  -ldflags="-s -w -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
  -o /shortener .

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /shortener .

RUN mkdir -p /data

EXPOSE 8080

CMD ["/app/shortener"]
