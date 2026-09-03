FROM node:22-alpine AS frontend
WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
ARG VITE_ONLYOFFICE_URL=http://localhost:8080
ENV VITE_ONLYOFFICE_URL=$VITE_ONLYOFFICE_URL
RUN npm run build

FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main ./cmd/app

FROM ubuntu:22.04
ENV GO_ENV=production \
    StaticDir=/app/frontend/dist
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=frontend /frontend/dist ./frontend/dist
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
EXPOSE 3053
CMD ["./main"]
