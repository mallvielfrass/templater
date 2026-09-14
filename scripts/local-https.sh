#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

if ! command -v mkcert >/dev/null 2>&1; then
  echo "mkcert is not installed" >&2
  exit 1
fi

mkcert -install || true
mkdir -p certs
mkcert -cert-file certs/templater.local.pem -key-file certs/templater.local-key.pem templater.local
cp "$(mkcert -CAROOT)/rootCA.pem" certs/rootCA.pem

if ! grep -qE '(^|[[:space:]])templater\.local([[:space:]]|$)' /etc/hosts; then
  if sudo -n true 2>/dev/null; then
    echo "127.0.0.1 templater.local" | sudo tee -a /etc/hosts >/dev/null
  else
    echo "add to /etc/hosts: 127.0.0.1 templater.local" >&2
  fi
fi

echo "https://templater.local:8443"
