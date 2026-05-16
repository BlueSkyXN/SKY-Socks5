# app/config navigation card

`app/config` owns CLI flag parsing, defaults, aliases, URL validation, timeout/concurrency validation, and output path resolution. Read this card before changing any flag, default, alias, help text, or output path rule. Key files: `config.go`, `config_test.go`.

## Local invariants

- `DefaultSourceFile` is `configs/sources.txt`.
- Default probe behavior is Cloudflare trace through `https://www.cloudflare.com/cdn-cgi/trace`.
- Compatibility aliases are part of the CLI surface: `-proxy` for `-fetch-proxy`, `-concurrency` for `-validate-concurrency`, and `-meta-output` for `-report-output`.
- All five output paths must resolve to distinct file paths: raw, unique, validated TXT, validated CSV, and JSON report.
- Timeouts and validation concurrency must be greater than zero.
- `-probe-url` must be HTTP or HTTPS; `-fetch-proxy` must be a supported proxy URL when provided.

## Local rules

- Update `Usage()` and tests in the same change when adding, renaming, or deleting a flag.
- Keep path cleaning centralized through existing helpers unless a caller needs different semantics.
- Preserve helpful error wrapping; pipeline surfaces these errors directly to CLI users.

## Do not

- Do not silently remove compatibility aliases.
- Do not allow output files to collide; collisions can overwrite artifacts from the same run.
- Do not add environment-variable configuration unless the user asks for it and tests cover precedence.

## Validation

- `go test ./app/config`
- `go run . -h` after help text or flag registration changes.
