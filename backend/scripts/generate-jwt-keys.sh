#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KEYS_DIR="${SCRIPT_DIR}/../dev-keys"

mkdir -p "$KEYS_DIR"

if [[ -f "$KEYS_DIR/jwt-private.pem" && -f "$KEYS_DIR/jwt-public.pem" ]]; then
    echo "JWT keys already exist at $KEYS_DIR"
    exit 0
fi

openssl genrsa -out "$KEYS_DIR/jwt-private.pem" 2048
openssl rsa -in "$KEYS_DIR/jwt-private.pem" -pubout -out "$KEYS_DIR/jwt-public.pem"

echo "JWT keys generated at $KEYS_DIR"
