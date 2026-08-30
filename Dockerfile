# Multi-Stage Dockerfile for nE Autonomous Personal Music Server

# Stage 1: Build React 19 Frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build Standalone Go Binary
FROM golang:alpine AS backend-builder
WORKDIR /app
ENV GOTOOLCHAIN=auto
RUN apk add --no-cache git gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy compiled frontend into web/dist for embedding
COPY --from=frontend-builder /app/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/ne ./cmd/ne

# Stage 3: Minimal Production Runtime
FROM alpine:3.21
RUN apk add --no-cache ffmpeg ca-certificates tzdata
WORKDIR /app

# Create standard directories
RUN mkdir -p /config /data /cache/artwork /cache/transcode /music /app/music

# Copy compiled backend binary
COPY --from=backend-builder /app/ne /usr/local/bin/ne

# Copy bundled music collection, database snapshot, and config keys into runtime container
COPY music/ /music/
COPY music/ /app/music/
COPY data/ /data/
COPY config/ /config/

EXPOSE 4533

ENV NE_PORT=4533 \
    NE_HOST=0.0.0.0 \
    NE_CONFIG_DIR=/config \
    NE_DATA_DIR=/data \
    NE_CACHE_DIR=/cache \
    NE_MUSIC_DIR=/music \
    NE_DB_PATH=/data/ne.db \
    NE_JWT_SECRET="ne-audio-streaming-production-secret-key-32b" \
    NE_STREAMING_MAX_CONCURRENT_TRANSCODES=1

ENTRYPOINT ["/usr/local/bin/ne"]
