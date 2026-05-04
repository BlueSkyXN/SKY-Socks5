# Proxy Sources

The default source list lives at `configs/sources.txt`.

Each non-empty, non-comment line should be an HTTP or HTTPS URL that returns newline-delimited SOCKS5 proxy candidates. The parser accepts common line formats such as:

- `host:port`
- `socks5://host:port`
- `socks5://user:pass@host:port`
- `[ipv6-address]:port`

## Current source policy

- Prefer SOCKS5-specific endpoints over mixed protocol lists.
- Prefer plain text endpoints over JSON or CSV because the CLI consumes one proxy candidate per line.
- Keep source additions small and reviewable.
- Verify a new source with a short run using a temporary source file before merging.

## Proxifly

Proxifly's SOCKS5 list is included through its text endpoint:

```text
https://cdn.jsdelivr.net/gh/proxifly/free-proxy-list@main/proxies/protocols/socks5/data.txt
```

This endpoint returns `socks5://host:port` lines, which are supported by the parser.
