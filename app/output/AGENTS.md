# app/output navigation card

`app/output` writes deterministic text, CSV, and JSON artifacts to disk. Read this card before changing file creation, sorting, deduplication, newline policy, CSV serialization, JSON formatting, or permissions. Key files: `output.go`; callers live in `app/pipeline`.

## Local invariants

- Text outputs trim blank lines, sort records, and end with a trailing newline when non-empty.
- Deduplication is caller-selected; raw output can preserve duplicate observations while unique and validated outputs dedupe.
- JSON output is indented with a trailing newline.
- CSV output uses Go's `encoding/csv` writer.
- Parent directories are created with `0755`, and files are written with `0644`.

## Local rules

- Keep write errors wrapped with enough path/context for CLI users.
- Check pipeline tests when changing writer semantics; artifact content expectations usually live there.
- If output paths or file names change, coordinate with `app/config`, `app/pipeline`, README, and docs.

## Do not

- Do not write outside the caller-resolved path.
- Do not remove deterministic sorting without updating comparison and reporting expectations.
- Do not swallow file-system errors.

## Validation

Use root validation commands. Add package tests if changing writer edge cases; run `go test ./app/output ./app/pipeline` for output behavior changes.
