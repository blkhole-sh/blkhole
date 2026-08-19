# syntax=docker/dockerfile:1

# =============================================================================
# Stage 1: Build Frontend (SolidJS)
# =============================================================================
FROM oven/bun:1.3.9 AS web

WORKDIR /app/web

# Copy dependencies first for better layer caching
COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile

# Build static frontend assets
COPY web/ ./
ARG VITE_BLKHOLE_EXTENSION_CHROMIUM_URL
ARG VITE_BLKHOLE_EXTENSION_FIREFOX_URL
ARG VITE_BLKHOLE_EXTENSION_SAFARI_URL
ENV VITE_BLKHOLE_EXTENSION_CHROMIUM_URL=$VITE_BLKHOLE_EXTENSION_CHROMIUM_URL
ENV VITE_BLKHOLE_EXTENSION_FIREFOX_URL=$VITE_BLKHOLE_EXTENSION_FIREFOX_URL
ENV VITE_BLKHOLE_EXTENSION_SAFARI_URL=$VITE_BLKHOLE_EXTENSION_SAFARI_URL
RUN bun run build

# =============================================================================
# Stage 2: Build Backend (Go)
# =============================================================================
FROM golang:1.26.4 AS builder

WORKDIR /app

# Copy dependencies first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source and frontend assets
COPY . .
COPY --from=web /app/static ./static

# CGO_ENABLED=1 for SQLite, -ldflags strips debug info, -trimpath for reproducibility
RUN CGO_ENABLED=1 go build -ldflags='-s -w' -trimpath -o blkhole .

# =============================================================================
# Stage 3: Runtime (Distroless)
# =============================================================================
# hadolint ignore=DL3006
FROM gcr.io/distroless/base-debian13

COPY --from=builder /app/blkhole /blkhole

EXPOSE 80 8080 853 853/udp
VOLUME /root/.config/blkhole  # Persistent config and database

ENTRYPOINT ["/blkhole"]
