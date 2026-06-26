# RunComfy CLI を Linux コンテナで動かすイメージ。
# Windows ホストでは Docker Desktop 経由で実行する。
#
# ビルド例:
#   docker compose build
#
# 実行例:
#   docker compose run --rm runcomfy --version

FROM debian:bookworm-slim

ARG RUNCOMFY_VERSION=v0.1.2

ENV DEBIAN_FRONTEND=noninteractive \
    PATH=/usr/local/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        tar \
    && rm -rf /var/lib/apt/lists/*

# install.sh が不安定なため GitHub Releases から直接取得（公式バイナリ + SHA-256）。
RUN set -eux; \
    ARCH="$(uname -m)"; \
    case "$ARCH" in \
        x86_64) TARGET="x86_64-unknown-linux-gnu" ;; \
        aarch64) TARGET="aarch64-unknown-linux-gnu" ;; \
        *) echo "unsupported architecture: $ARCH" >&2; exit 1 ;; \
    esac; \
    BASE_URL="https://github.com/runcomfy-com/runcomfy-cli/releases/download/${RUNCOMFY_VERSION}"; \
    ARCHIVE="runcomfy-${TARGET}.tar.gz"; \
    curl -fsSL "${BASE_URL}/${ARCHIVE}" -o "/tmp/${ARCHIVE}"; \
    curl -fsSL "${BASE_URL}/${ARCHIVE}.sha256" -o "/tmp/${ARCHIVE}.sha256"; \
    cd /tmp; \
    sha256sum -c "${ARCHIVE}.sha256"; \
    tar -xzf "${ARCHIVE}" -C /usr/local/bin; \
    chmod +x /usr/local/bin/runcomfy; \
    runcomfy --version; \
    rm -f "/tmp/${ARCHIVE}" "/tmp/${ARCHIVE}.sha256"

WORKDIR /workspace

ENTRYPOINT ["runcomfy"]
