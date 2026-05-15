#!/usr/bin/env bash
set -e

ROLE="${ROLE:-all-in-one}"

# Generate self-signed cert if SSL vars are not set
if [ -z "${UVICORN_SSL_CERTFILE:-}" ] || [ -z "${UVICORN_SSL_KEYFILE:-}" ]; then
    CERT_DIR="${SSL_CERT_DIR:-/var/lib/pasarguard/certs}"
    mkdir -p "$CERT_DIR"
    if [ ! -f "$CERT_DIR/ssl_cert.pem" ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] Generating self-signed certificate..."
        openssl req -x509 -newkey rsa:4096 \
            -keyout "$CERT_DIR/ssl_key.pem" \
            -out "$CERT_DIR/ssl_cert.pem" \
            -days 36500 -nodes \
            -subj "/CN=localhost" \
            -addext "subjectAltName=DNS:localhost,IP:127.0.0.1" 2>/dev/null
    fi
    export UVICORN_SSL_CERTFILE="$CERT_DIR/ssl_cert.pem"
    export UVICORN_SSL_KEYFILE="$CERT_DIR/ssl_key.pem"
    export UVICORN_SSL_CA_TYPE="private"
fi

if [ "${ROLE}" = "node" ]; then
    exec python node_worker.py
elif [ "${ROLE}" = "scheduler" ]; then
    exec python scheduler_worker.py
else
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Starting ${ROLE}..."
    python -m alembic upgrade head
    exit_code=$?

    if [ $exit_code -ne 0 ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: Database migrations failed"
        exit 1
    fi

    exec python main.py
fi
