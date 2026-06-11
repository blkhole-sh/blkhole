# blkhole

A self-hosted DNS-over-HTTPS content blocker with schedule-based filtering,
per-device rules, and a web admin interface.

---

## Features

- DNS-over-HTTPS server with per-device routing
- Domain blocklists — import public lists or build your own
- Schedule-based blocking — block by time of day and day of week
- Allow/block rules per domain
- One-click mobileconfig for iOS and macOS
- Web admin interface
- Automatic HTTPS via Let's Encrypt
- Single binary — no runtime dependencies

---

## Installation

### go install

Requires Go 1.26+ and a C compiler (gcc or clang) for SQLite.

```sh
CGO_ENABLED=1 go install github.com/blkhole-sh/blkhole@latest
```

### Docker

**Standard (Distroless):**

Smallest image, minimal attack surface.

```sh
docker build -t blkhole .
docker run -d \
  -v blkhole-data:/root/.config/blkhole \
  -p 80:80 -p 443:443 \
  blkhole -d yourdomain.com -s $(openssl rand -hex 32)
```

**With Unbound (Self-hosted DNS resolver):**

Includes Unbound recursive DNS resolver.

```sh
docker build -f Dockerfile.unbound -t blkhole:unbound .
docker run -d \
  -v blkhole-data:/data \
  -p 53:53/udp -p 53:53/tcp -p 8080:8080 \
  blkhole:unbound
```

Flow: `Client → blkhole:53 → Unbound:5353 → Root DNS servers`

---

## Usage

```sh
# Production (HTTPS with autocert)
blkhole -d yourdomain.com -s $(openssl rand -hex 32)

# Local (plain HTTP)
blkhole -p 8080 -s $(openssl rand -hex 32)
```

### Flags

| Flag | Description | Default |
|---|---|---|
| `-d <domain>` | Production mode — HTTPS with autocert | — |
| `-p <port>` | Local mode — plain HTTP | — |
| `-u <host:port>` | Upstream DNS server | `1.1.1.1:53` |
| `-s <hex>` | JWT secret (hex-encoded, min 32 bytes) | — |

`-d` and `-p` are mutually exclusive. One is required.

> **Note:** Production mode requires ports 80 and 443 to be open and your domain's DNS pointing to the server for autocert to issue a certificate.

---

## Setup

1. Open the admin interface and sign in
2. **Devices** — add a device and download its mobileconfig to configure DoH on iOS or macOS. For other platforms use the DoH URL directly:
   ```
   https://yourdomain.com/{deviceHash}/dns-query
   ```
3. **Blocklists** — create a list by importing a public blocklist URL or adding domains manually
4. **Schedules** — create a schedule to define when blocking is active (time of day, days of week), then attach devices and lists to it

---

## Building from Source

```sh
cd web && bun install && bun run build && cd ..
go build -ldflags='-s -w' -trimpath -o blkhole .
```

Requires Bun 1.3+ for the frontend build step.

---

## Contributing

Contributions are welcome. Please open an issue before submitting a pull request for significant changes.

---

## License

AGPL-3.0 — see [LICENSE](LICENSE)
