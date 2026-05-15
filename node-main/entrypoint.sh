#!/bin/sh
set -e

# Generate self-signed TLS certs if missing
CERT_DIR=/var/lib/pg-node/certs
if [ ! -f "$CERT_DIR/ssl_cert.pem" ] || [ ! -f "$CERT_DIR/ssl_key.pem" ]; then
    echo "Generating self-signed TLS certificate..."
    mkdir -p "$CERT_DIR"
    openssl req -x509 -newkey rsa:4096 \
        -keyout "$CERT_DIR/ssl_key.pem" \
        -out "$CERT_DIR/ssl_cert.pem" \
        -days 36500 -nodes \
        -subj "/CN=localhost" \
        -addext "subjectAltName=DNS:localhost,IP:127.0.0.1" 2>/dev/null
    echo "TLS certificate generated."
fi

# Ensure generated config directory exists
mkdir -p /var/lib/pg-node/generated

exec ./main
