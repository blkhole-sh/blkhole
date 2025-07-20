FROM oven/bun:1-alpine AS builder

WORKDIR /app
COPY package.json bun.lockb ./
RUN bun install

COPY . .
RUN bun run build

FROM oven/bun:1-alpine
WORKDIR /app
COPY --from=builder /app/.output ./.output
COPY --from=builder /app/package.json ./

EXPOSE ${APP_PORT:-3000}
CMD ["bun", "start"]