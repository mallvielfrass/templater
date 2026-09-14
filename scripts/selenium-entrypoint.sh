#!/usr/bin/env bash
set -euo pipefail

if [ -f /certs/rootCA.pem ]; then
  sudo cp /certs/rootCA.pem /usr/local/share/ca-certificates/mkcert.crt
  sudo update-ca-certificates || true
  mkdir -p "${HOME}/.pki/nssdb"
  if [ ! -f "${HOME}/.pki/nssdb/cert9.db" ]; then
    touch /tmp/empty-nss-pass
    certutil -N -d "sql:${HOME}/.pki/nssdb" -f /tmp/empty-nss-pass >/dev/null 2>&1 || true
    rm -f /tmp/empty-nss-pass
  fi
  certutil -d "sql:${HOME}/.pki/nssdb" -D -n mkcert >/dev/null 2>&1 || true
  certutil -d "sql:${HOME}/.pki/nssdb" -A -t "C,," -n mkcert -i /certs/rootCA.pem
fi

exec /opt/bin/entry_point.sh "$@"
