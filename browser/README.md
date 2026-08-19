# blkhole browser extension

The extension mirrors the active block rules for a paired device and redirects only top-level browser navigations to a packaged block page. DNS blocking remains authoritative for all other requests.

## Build and test

```sh
bun test
bun run build
```

The build creates unpacked extensions in `build/chromium`, `build/firefox`, and `build/safari`. The Chromium bundle also runs in Edge and Brave. Release and store builds should set their trusted web app origin with `BLKHOLE_DEFAULT_API_ORIGIN=https://example.com bun run build`. The value must be a bare HTTPS origin. Without it the extension starts unconfigured and must be configured explicitly on its options page before the web app bridge responds.

Safari 16.4 or newer is required because the rules use Manifest V3 `declarativeNetRequest` domain conditions. `bun run package:safari` invokes Apple's converter to create an untracked Xcode wrapper with macOS and iOS/iPadOS targets. The invoking terminal must have macOS file access to the source directory. Code signing, provisioning profiles, final device testing, and App Store packaging still happen in Xcode.

## Pairing and permissions

The web app bridge listens for `BLKHOLE_EXTENSION_PING` and `BLKHOLE_EXTENSION_PAIR` messages from the page and echoes the supplied nonce. It answers only on the configured API origin. A self-hosted origin must first be saved explicitly on the extension's options page; changing it unpairs the extension and clears all rules.

`<all_urls>` is required so declarative rules can redirect any blocked top-level domain. The content bridge is loaded on HTTPS pages, but the background process rejects every origin except the configured API origin. HTTP origins are never accepted.

The bearer credential is stored in extension-local storage and sent only to the configured origin. Rules are refreshed every minute with `If-None-Match`. A `401` or `403` clears the credential and the dynamic rules. Rule replacement is submitted as one atomic declarativeNetRequest update. Domains are grouped into deterministic chunks so large lists do not consume one dynamic-rule slot per domain; rejected updates retain the previous rules and are never silently truncated.

"Clear local pairing" removes the credential and rules from this browser. Revoke the browser in the web app when the server-side credential must be invalidated as well.

The most recent top-level hostname per tab is kept in `storage.session` so a suspended Manifest V3 service worker can still label the packaged block page. Browsers without `storage.session` fall back to memory; Firefox's persistent background page and the immediate redirect handoff cover that compatibility path.
