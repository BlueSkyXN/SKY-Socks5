# app/proxy navigation card

`app/proxy` owns SOCKS5 candidate parsing, normalization, validation, metadata extraction, and deterministic sorting. Read this card before changing accepted input formats, scheme rules, CSV-like metadata handling, IPv6 handling, or sort behavior. Key files: `proxy.go`, `proxy_test.go`, `docs/sources.md`.

## Local invariants

- Accepted candidate forms include `host:port`, `socks5://host:port`, `socks5://user:pass@host:port`, CSV-like `socks5://host:port,country,city`, and bracketed IPv6 with a port.
- HTTP and HTTPS proxy records are not SOCKS5 candidates and must not be tested as SOCKS5.
- Bare IPv6 addresses must use bracket notation.
- Normalized addresses are `net.JoinHostPort(host, port)` with lower-cased hosts.
- Blank lines and full-line comments are structural text, not parse errors.
- Sorting must remain deterministic by host and numeric port.

## Local rules

- Add table-driven tests for every accepted or rejected input format change.
- Keep source metadata extraction limited and explicit; the main CSV only consumes source country and city.
- Check `docs/sources.md` when changing parsing assumptions for upstream source formats.

## Do not

- Do not add mixed-protocol support here by testing `http://` or `https://` records as SOCKS5.
- Do not treat high upstream candidate counts as validation quality.
- Do not change normalized output format without updating downstream tests and docs.

## Validation

- `go test ./app/proxy`
- `go test ./app/pipeline` when normalized address, metadata, or sorting semantics affect reports.
