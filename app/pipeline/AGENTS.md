# app/pipeline navigation card

`app/pipeline` orchestrates the full CLI run and converts source, fetch, parse, validation, output, and report data into user-facing artifacts. Read this card before changing runtime order, totals, per-source counts, CSV columns, dependency injection, or CLI summary output. Key files: `pipeline.go`, `pipeline_test.go`, `docs/result-fields.md`.

## Local invariants

- `RunCLI` parses config, calls `Run`, and prints one compact stdout summary.
- `Run` must continue through per-source fetch and parse errors so reports can explain partial failures.
- Raw, unique, validated TXT, validated CSV, and JSON report outputs are all written from one run.
- Raw and unique proxy lists are sorted deterministically before validation/output.
- Per-source `ValidProxies` is computed after global validation, counting a valid repeated proxy for every source that listed it.
- CSV header order is an output contract and must stay aligned with `docs/result-fields.md`.

## Local rules

- Use `Dependencies` for tests instead of introducing live network or wall-clock dependencies.
- When changing count semantics, update pipeline tests and CD summary assumptions.
- When changing CSV/report fields, read `docs/result-fields.md` first and update `app/report` together.

## Do not

- Do not make one failed upstream source abort the entire aggregate run unless explicitly requested.
- Do not write artifacts before all required in-memory metadata for the run is available.
- Do not add new stdout formats without considering scripts that may parse the current summary.

## Validation

- `go test ./app/pipeline`
- `go test ./app/report ./app/output ./app/pipeline` for artifact/schema changes.
- `go test ./...` before handoff for runtime flow changes.
