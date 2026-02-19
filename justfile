dev:
    cd web && bun dev &
    air

build:
    cd web && bun run build
    go build -ldflags='-s -w' -trimpath -o blkhole .
