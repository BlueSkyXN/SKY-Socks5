# app/validate navigation card

`app/validate` tests deduplicated SOCKS5 proxies through an HTTP probe URL and extracts compact Cloudflare trace evidence. Read this card before changing live validation, concurrency, accepted status handling, trace parsing, or result fields. Key files: `validate.go`, `validate_test.go`, `docs/result-fields.md`.

## Local invariants

- Validation uses `socks5://<address>` as the HTTP client proxy.
- `ProbeURL`, positive `Timeout`, positive `Concurrency`, and non-empty `ValidStatuses` are required.
- A proxy is reachable only when the probe request completes and returns an accepted HTTP status.
- Cloudflare trace is parsed as `key=value` lines; CSV-facing fields are limited to compact evidence such as exit IP, country, colo, HTTP, and non-default flags.
- Full raw trace data belongs in the JSON report, not expanded blindly into CSV columns.

## Local rules

- Keep validation context-aware and bounded by caller-provided timeout/concurrency.
- Use the `Probe` hook or small parser tests for unit coverage; do not require live public proxies in tests.
- Read `docs/result-fields.md` before adding or removing trace-derived fields.

## Do not

- Do not present local validation output as CD-quality evidence.
- Do not add ASN, ISP, city, region, or risk-score fields from Cloudflare trace; the docs state Cloudflare trace does not provide them.
- Do not make validation fail the whole run because one proxy is unreachable.

## Validation

- `go test ./app/validate`
- `go test ./app/pipeline` when result semantics feed CSV or JSON outputs.
