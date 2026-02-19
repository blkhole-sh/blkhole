# Stage 1: Build frontend
FROM oven/bun:1.3.9 AS web
WORKDIR /app/web
COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile
COPY web/ ./
RUN bun run build

# Stage 2: Build Go binary
FROM golang:1.26.0 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /app/static ./static
RUN CGO_ENABLED=1 go build -ldflags='-s -w' -trimpath -o blkhole .

# Stage 3: Runtime
# hadolint ignore=DL3006
FROM gcr.io/distroless/base-debian13
COPY --from=builder /app/blkhole /blkhole
EXPOSE 80 443
VOLUME /root/.config/blkhole
ENTRYPOINT ["/blkhole"]
