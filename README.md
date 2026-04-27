# SKY-Socks5

SKY-Socks5 is a Go-first CLI for collecting public SOCKS5 proxy lists, normalizing them, removing duplicates, validating reachability, and publishing deterministic output files for downstream use.

> Public proxies are noisy and untrusted. Use this project for research, testing, and automation at your own risk.

## What the CLI does

The Go pipeline now owns the primary runtime:

1. Load upstream source URLs from `urls.txt` or a custom source file.
2. Fetch candidate SOCKS5 entries from each upstream list.
3. Normalize and parse proxy addresses.
4. Deduplicate the aggregated results.
5. Validate reachability against a probe URL.
6. Write text outputs plus a machine-readable JSON report.

## Project layout

- `cmd/sky-socks5/` - main Go CLI entrypoint.
- `internal/` - fetch, source loading, parsing, validation, output, and reporting packages.
- `main.go` - thin wrapper around the Go CLI pipeline.
- `urls.txt` - default list of proxy source URLs.

## Requirements

- Go 1.22+

## Quick start

Run the test suite:

```bash
go test ./...
```

Run the CLI with default settings:

```bash
go run ./cmd/sky-socks5
```

Build a reusable binary:

```bash
go build -o bin/sky-socks5 ./cmd/sky-socks5
./bin/sky-socks5
```

The CLI prints a one-line summary with counts and the paths of the generated artifacts.

## Default output files

By default the CLI writes files into the current directory. Use `-output-dir` to place them elsewhere.

| File | Description |
| --- | --- |
| `raw_proxies.txt` | All successfully parsed proxy addresses collected from every source, sorted, before deduplication. |
| `unique_proxies.txt` | Deduplicated proxy list. |
| `validated_proxies.txt` | Deduplicated proxies that passed live validation. |
| `proxy_report.json` | Metadata report with totals, per-source stats, timeouts, probe URL, errors, and output paths. |

## Basic usage

Use the defaults:

```bash
go run ./cmd/sky-socks5
```

Write artifacts into a separate directory:

```bash
go run ./cmd/sky-socks5 -output-dir generated
```

Use a custom source file and validation settings:

```bash
go run ./cmd/sky-socks5 \
  -source-file custom-urls.txt \
  -output-dir out \
  -probe-url https://example.com \
  -validate-concurrency 20 \
  -validate-timeout 15s
```

## Configuration and flags

Common flags:

- `-source-file` - path to the file that lists upstream proxy URLs.
- `-output-dir` - directory for generated artifacts.
- `-raw-output`, `-unique-output`, `-validated-output`, `-meta-output` - override output filenames or paths.
- `-probe-url` - HTTP or HTTPS URL used when validating proxies.
- `-fetch-timeout` - timeout for downloading source lists.
- `-validate-timeout` - timeout per proxy validation attempt.
- `-validate-concurrency` - maximum number of concurrent validation workers.

For the full generated help text:

```bash
go run ./cmd/sky-socks5 -h
```

## Source file format

`urls.txt` is a newline-delimited list of `http://` or `https://` URLs. Blank lines and lines starting with `#` are ignored.

## GitHub automation

- **Go CI** runs `go test ./...`.
- **Publish Proxy Artifacts** runs the Go CLI on a schedule or by manual dispatch and uploads the generated outputs.
- **Release** builds archives from `./cmd/sky-socks5` for tagged releases.

## Deprecated Python scripts

The legacy Python scripts (`Get_Socks5_List.py`, `Test_Sock5_List.py`, and `S5HUB.py`) are still kept in the repository as deprecated compatibility artifacts. They are no longer the primary runtime, CI path, or release target for this project.
