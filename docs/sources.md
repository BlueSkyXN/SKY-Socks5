# Proxy Sources

The default source list lives at `configs/sources.txt`.

Each non-empty, non-comment line should be an HTTP or HTTPS URL that returns SOCKS5 proxy candidates. Newline-delimited lists are preferred, but whitespace-delimited records are also supported for sources that publish many candidates on one line. The parser accepts common line formats such as:

- `host:port`
- `socks5://host:port`
- `socks5://user:pass@host:port`
- `socks5://host:port,FR,Paris`
- `[ipv6-address]:port`

## Current source policy

- Prefer SOCKS5-specific endpoints over mixed protocol lists.
- Prefer text endpoints over JSON. CSV-like records are accepted when the first field is the proxy URL and following fields provide source metadata such as country or city.
- Keep source additions small and reviewable.
- Verify a new source with a short run using a temporary source file before merging.
- Country-scoped Proxifly files can contain mixed protocols. The CLI accepts SOCKS5 candidates and rejects `http://` or `https://` proxy records instead of testing them as SOCKS5.
- A high upstream candidate count is not the same as a high validated count. Validation requires an end-to-end SOCKS5 connection to the probe URL within the configured timeout and an accepted HTTP status.

## Proxifly

Proxifly's SOCKS5 list is included through its text endpoint:

```text
https://cdn.jsdelivr.net/gh/proxifly/free-proxy-list@main/proxies/protocols/socks5/data.txt
```

This endpoint returns `socks5://host:port` lines, which are supported by the parser.

Proxifly country endpoints, for example `proxies/countries/FR/data.txt` or `data.csv`, are useful for country-specific smoke runs, but they are broader than the default SOCKS5 protocol feed. Use their SOCKS5 records when the goal is this repository's `validated_proxies.txt` and `validated_proxies.csv` outputs.
