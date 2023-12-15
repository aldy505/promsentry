FROM golang:1.21-bookworm AS builder

WORKDIR /app

COPY . .

RUN go build -o promsentry -ldflags="-s -w" ./cmd

FROM debian:bookworm-slim AS runtime

RUN apt-get update && apt-get install -y curl ca-certificates

COPY --from=builder /app/promsentry /usr/local/bin/promsentry

EXPOSE 3000

HEALTHCHECK --start-period=10s --interval=60s \
    CMD curl -f http://localhost:3000/ || exit 1

CMD ["/usr/local/bin/promsentry"]
