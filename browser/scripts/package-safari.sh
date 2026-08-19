#!/bin/sh
set -eu

if ! command -v xcrun >/dev/null 2>&1; then
	echo "xcrun is required to create the Safari Xcode wrapper" >&2
	exit 1
fi

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
EXTENSION_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
SAFARI_BUILD="$EXTENSION_ROOT/build/safari"
SAFARI_PROJECT="$EXTENSION_ROOT/build/safari-native"

cd "$EXTENSION_ROOT"
bun run build:safari
rm -rf "$SAFARI_PROJECT"
xcrun safari-web-extension-converter "$SAFARI_BUILD" \
	--project-location "$SAFARI_PROJECT" \
	--app-name blkhole \
	--bundle-identifier sh.blkhole.browser \
	--swift \
	--copy-resources \
	--no-prompt \
	--no-open
