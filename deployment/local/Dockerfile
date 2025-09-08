FROM golang:1.24.1-alpine AS builder
WORKDIR /app
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates
COPY . .
ARG BUILD_TARGET
RUN go build -o /app/bin/app $BUILD_TARGET

FROM scratch
COPY --from=builder /app/bin/app /
COPY --from=builder /app/internal/config /internal/config
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/app"]
