FROM golang:1.24.1-bullseye AS builder
WORKDIR /app
RUN apt-get update 
RUN apt-get install -y git ca-certificates
RUN apt-get install -y build-essential librdkafka-dev ca-certificates git make
RUN update-ca-certificates 
COPY . .
ARG BUILD_TARGET
RUN CGO_ENABLED=1 go build -o /app/bin/app $BUILD_TARGET

FROM golang:1.24.1-bullseye AS final
WORKDIR /
COPY --from=builder /app/bin/app /
COPY --from=builder /app/internal/config /internal/config
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
RUN chmod 755 -R /app && chmod -R 755 /internal
ENTRYPOINT ["/app"]
