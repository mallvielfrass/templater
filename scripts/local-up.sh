#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

if [[ ! -f certs/templater.local.pem || ! -f certs/rootCA.pem ]]; then
  echo "run scripts/local-https.sh first" >&2
  exit 1
fi

if ! grep -qE '(^|[[:space:]])templater\.local([[:space:]]|$)' /etc/hosts; then
  echo "add to /etc/hosts: 127.0.0.1 templater.local" >&2
fi

docker network inspect templater-backend-network >/dev/null 2>&1 || docker network create templater-backend-network
docker volume inspect templater-backend-data >/dev/null 2>&1 || docker volume create templater-backend-data

docker build -t templater-onlyoffice -f Dockerfile.onlyoffice .
docker build -t templater-backend --build-arg VITE_ONLYOFFICE_URL=https://templater.local:8443/ds .
docker build -t templater-selenium -f Dockerfile.selenium .

docker rm -f templater-onlyoffice templater-backend templater-nginx templater-selenium >/dev/null 2>&1 || true

docker run -d --name templater-onlyoffice \
  --network templater-backend-network \
  --network-alias onlyoffice \
  --network-alias documentserver \
  --add-host host.docker.internal:host-gateway \
  --add-host templater.local:host-gateway \
  -e JWT_ENABLED=true \
  -e JWT_SECRET=onlyoffice-dev-secret \
  -e JWT_HEADER=Authorization \
  -e ALLOW_PRIVATE_IP_ADDRESS=true \
  -e NODE_EXTRA_CA_CERTS=/certs/rootCA.pem \
  -e SSL_CERT_FILE=/certs/rootCA.pem \
  -v "$root/certs/rootCA.pem:/certs/rootCA.pem:ro" \
  -v "$root/scripts/onlyoffice-with-ca.sh:/onlyoffice-with-ca.sh:ro" \
  --entrypoint /bin/bash \
  templater-onlyoffice /onlyoffice-with-ca.sh

docker run -d --name templater-backend \
  --network templater-backend-network \
  --network-alias templater-backend \
  --add-host host.docker.internal:host-gateway \
  --add-host templater.local:host-gateway \
  -e HttpPort=3053 \
  -e BadgerDBPath=/storage/badger \
  -e JWTSecret=dev-jwt-secret-change-me \
  -e PublicBaseURL=https://templater.local:8443 \
  -e OnlyOfficeJWTSecret=onlyoffice-dev-secret \
  -e OnlyOfficeURL=http://onlyoffice \
  -e CORSOrigins=https://templater.local:8443 \
  -e GO_ENV=development \
  -e StaticDir=/app/frontend/dist \
  -v templater-backend-data:/storage \
  templater-backend

docker run -d --name templater-nginx \
  --network templater-backend-network \
  --add-host host.docker.internal:host-gateway \
  --add-host templater.local:host-gateway \
  -p 8088:80 \
  -p 8443:443 \
  -v "$root/nginx/nginx.local.conf:/etc/nginx/nginx.conf:ro" \
  -v "$root/certs:/etc/nginx/certs:ro" \
  nginx:1.27-alpine

docker run -d --name templater-selenium \
  --network templater-backend-network \
  --add-host templater.local:host-gateway \
  --shm-size 2g \
  -p 4444:4444 \
  -v "$root/certs/rootCA.pem:/certs/rootCA.pem:ro" \
  -v "$root/test.xlsx:/fixtures/test.xlsx:ro" \
  -v "$root/test.docx:/fixtures/test.docx:ro" \
  templater-selenium

echo "https://templater.local:8443"
