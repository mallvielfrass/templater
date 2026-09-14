#!/bin/bash
set -euo pipefail

if [ -f /certs/rootCA.pem ]; then
  cp /certs/rootCA.pem /usr/local/share/ca-certificates/mkcert.crt
  update-ca-certificates || true
fi

exec /app/ds/run-document-server.sh
