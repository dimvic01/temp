FROM golang:1.26 AS builder
WORKDIR /src
COPY . .
RUN echo "appuser:x:1001:1001:App User:/:/sbin/nologin" > /etc/passwd-scratch && \
    echo "appgroup:x:1001:" > /etc/group-scratch && \
    go get github.com/prometheus/client_golang/prometheus && \
    go get github.com/prometheus/client_golang/prometheus/promauto && \
    go get github.com/prometheus/client_golang/prometheus/promhttp && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /bin/myapp1-api .

#FROM debian:bookworm-slim
#FROM ubuntu/dotnet-deps:8.0-24.04_stable
FROM scratch

EXPOSE 8080
WORKDIR /app

COPY --from=builder /etc/passwd-scratch /etc/passwd
COPY --from=builder /etc/group-scratch /etc/group
#    COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/myapp1-api /app/myapp1-api

#RUN groupadd -g 1001 appgroup && \
#    useradd -u 1001 -g appgroup -m -s /bin/bash appuser && \
#    chown -R appuser:appgroup /app

USER 1001
#USER appuser

HEALTHCHECK --interval=60s --timeout=5s --start-period=5s --retries=3 \
  CMD ["./myapp1-api", "--healthcheck"]

CMD ["./myapp1-api"]
