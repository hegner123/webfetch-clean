# Build stage
FROM golang:1-bookworm AS build

WORKDIR /src

# Cache dependency downloads
COPY go.mod go.sum ./
RUN go mod download

# Build binary
COPY . .
RUN go build -ldflags="-s -w" -o /webfetch-clean

# Runtime stage
FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        chromium \
        curl && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user and data directories
RUN useradd -r -u 1000 -m webfetch && \
    mkdir -p /data/db /data/files && \
    chown -R webfetch:webfetch /data

COPY --from=build /webfetch-clean /usr/local/bin/webfetch-clean

USER webfetch

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["webfetch-clean"]
CMD ["--http", "0.0.0.0:8080", "--db", "/data/db/webfetch.db"]
