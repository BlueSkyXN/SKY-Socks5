# SKY-Socks5

SKY-Socks5 is a small Go CLI for collecting public SOCKS5 proxy lists, normalizing `host:port` entries, removing duplicates, validating reachability, and publishing reproducible artifact files.

Public proxies are noisy and untrusted. Use this project for research, testing, and automation at your own risk.

## What it does

1. Loads source URLs from `urls.txt` or a custom source file.
2. Downloads each upstream proxy list.
3. Parses and normalizes SOCKS5 proxy addresses.
4. Writes parsed and deduplicated proxy files.
5. Validates proxies through a configurable probe URL.
6. Writes a JSON report with source, parse, validation, and output metadata.

## Requirements

- Go 1.22+

## Quick start

Run tests:

```bash
go test ./...
```

Show CLI help:

```bash
go run ./cmd/sky-socks5 -h
```

Generate artifacts with the default source file:

```bash
go run ./cmd/sky-socks5 -output-dir generated
```

Build a local binary:

```bash
go build -o bin/sky-socks5 ./cmd/sky-socks5
./bin/sky-socks5 -output-dir generated
```

## Output files

| File | Description |
| --- | --- |
| `raw_proxies.txt` | Parsed proxy addresses before deduplication. |
| `unique_proxies.txt` | Deduplicated proxy addresses. |
| `validated_proxies.txt` | Deduplicated proxies that passed live validation. |
| `proxy_report.json` | Machine-readable run report with source and validation metadata. |

## Common flags

- `-source-file` - newline-delimited source URL file. Defaults to `urls.txt`.
- `-output-dir` - directory for generated artifacts.
- `-fetch-proxy` / `-proxy` - optional proxy used when fetching source lists.
- `-probe-url` - HTTP or HTTPS URL used to validate candidate proxies.
- `-valid-statuses` - comma-separated HTTP statuses accepted during validation. Defaults to `200`.
- `-fetch-timeout` - timeout for fetching each source.
- `-validate-timeout` - timeout for each proxy validation request.
- `-validate-concurrency` / `-concurrency` - maximum validation workers.
- `-raw-output`, `-unique-output`, `-validated-output`, `-report-output` - override output filenames or paths.

Example:

```bash
go run ./cmd/sky-socks5 \
  -source-file urls.txt \
  -output-dir generated \
  -fetch-proxy socks5://127.0.0.1:1080 \
  -probe-url https://one.one.one.one \
  -valid-statuses 200,204 \
  -validate-concurrency 100
```

## GitHub automation

- **Go CI** runs `go test ./...`.
- **Publish Proxy Artifacts** is manual-only and uploads the generated files from `generated/`.
- **Release** builds cross-platform archives from `./cmd/sky-socks5` for tags or manual dispatch.

## Governance

The repository is intended to use pull requests, required checks, and CODEOWNERS review for changes to `main`. The primary CLI entrypoint is `./cmd/sky-socks5`; root-level runtime entrypoints and generated artifacts should not be committed.
