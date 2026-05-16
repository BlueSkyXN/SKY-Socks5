# Repository agent instructions

## Purpose

SKY-Socks5 is a small Go CLI for collecting public SOCKS5 proxy lists, parsing and normalizing `host:port` candidates, validating reachability through a probe URL, and writing deterministic TXT, CSV, and JSON artifacts.

Public proxies are noisy and untrusted. Treat local runs as development evidence only; the repository's live proxy validation signal belongs to the GitHub Actions CD workflow.

## Codex startup behavior

- Codex normally starts from the repository root. This file is the root router for that workflow.
- Subdirectory `AGENTS.md` files are local navigation cards. They are not guaranteed to be in startup context when Codex starts at the root.
- Before editing a path with `Local AGENTS.md = Yes` in the directory map, read the local card with `cat <path>/AGENTS.md`.
- If multiple nested cards apply, read them from shallow to deep before making changes.
- If a future `AGENTS.override.md` appears in a target directory, stop and ask the user how to handle that override before editing ordinary `AGENTS.md` there.

## Directory map

| Path | Responsibility | Local AGENTS.md | Read when |
|---|---|---:|---|
| `main.go` | CLI process entrypoint; delegates to `app/pipeline`. | No | Read before changing process exit, stdout/stderr behavior, or CLI startup. |
| `go.mod` | Go module declaration; currently Go 1.22 with no third-party requirements. | No | Read before changing Go version or adding dependencies. |
| `app/` | Application packages for config, source loading, fetch, proxy parsing, validation, output, reporting, and orchestration. | Yes | Read before any Go code change under `app/`, especially package boundary changes. |
| `app/config/` | CLI flags, defaults, aliases, URL checks, output path resolution. | Yes | Read before changing flags, defaults, validation, output paths, or help text. |
| `app/source/` | Loads source URLs from newline-delimited files. | Yes | Read before changing source file parsing, comment handling, or empty-source behavior. |
| `app/fetch/` | Concurrently downloads upstream source lists and records HTTP/source metadata. | Yes | Read before changing fetch concurrency, proxy use, timeout behavior, HTTP status handling, or response line parsing. |
| `app/proxy/` | Parses, normalizes, deduplicates, validates, and sorts SOCKS5 candidate addresses. | Yes | Read before changing accepted input formats, scheme policy, metadata extraction, IPv6 handling, or sort order. |
| `app/validate/` | Tests unique SOCKS5 proxies through an HTTP probe URL and extracts Cloudflare trace fields. | Yes | Read before changing live validation, concurrency, accepted status behavior, trace parsing, or result semantics. |
| `app/output/` | Writes deterministic TXT, CSV, and JSON artifact files. | Yes | Read before changing output writer behavior, file permissions, sorting, newline policy, or CSV/JSON serialization. |
| `app/report/` | Defines JSON report and validated proxy record schema. | Yes | Read before changing report fields, JSON names, totals, or output schema compatibility. |
| `app/pipeline/` | Orchestrates the full run and maps source/fetch/parse/validate output into artifacts and metadata. | Yes | Read before changing runtime flow, totals, per-source counts, CSV rows, or dependency injection. |
| `configs/` | Versioned runtime defaults, especially `configs/sources.txt`. | Yes | Read before adding, removing, or reordering default upstream proxy source URLs. |
| `docs/` | Human documentation and project policy notes. | No | Read the relevant doc before changing behavior it describes; do not create local rules here unless docs gain separate workflow constraints. |
| `.github/` | Repository control plane: CODEOWNERS and GitHub Actions workflows. | Yes | Read before changing CODEOWNERS, permissions, workflow triggers, or automation policy. |
| `.github/workflows/` | CI, build, and CD workflow definitions. | Yes | Read after `.github/AGENTS.md` before editing any workflow file. |
| `.codex/` | Local workspace state if present. It is not project source. | No | Do not rely on it for repository behavior unless files are explicitly tracked and requested. |

## On-demand cat protocol

Before editing files under a directory that has a local `AGENTS.md`, read that file first.

Examples:

```bash
cat app/AGENTS.md
cat app/proxy/AGENTS.md
```

For `.github/workflows/`, read both:

```bash
cat .github/AGENTS.md
cat .github/workflows/AGENTS.md
```

If a local card points to a specific doc, read that doc before editing behavior. For example, source-list work should read `docs/sources.md`, and report/CSV schema work should read `docs/result-fields.md`.

## Commands

| Command | Purpose | Scope | Sandbox notes |
|---|---|---|---|
| `go test ./...` | Run all Go package tests. | repo | OK in a normal local sandbox. |
| `go vet ./...` | Static Go checks recommended by project docs. | repo | OK in a normal local sandbox. |
| `go run . -h` | Print CLI help and verify flag registration. | root CLI | OK; no network expected. |
| `go build -o bin/sky-socks5 .` | Build a local binary as shown in README. | root CLI | OK; writes `bin/`, do not commit binary artifacts. |
| `go run . -output-dir generated` | Run the default source fetch, parse, validation, and artifact write flow. | root CLI | Requires external network and public proxies; writes `generated/`; do not use local reachability as CD evidence. |
| `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/sky-socks5 .` | Representative cross-build from the build workflow. | root CLI | OK; writes `dist/`, do not commit artifacts. |
| `actionlint .github/workflows/*.yml` | Lint GitHub Actions workflow syntax. | `.github/workflows/` | Requires `actionlint` installed locally; not vendored by this repo. |

## Global rules

- Go version is defined by `go.mod`. Keep code compatible with that version unless the user asks for an upgrade.
- Prefer Go standard library facilities. This module currently has no third-party dependencies; add dependencies only when the standard library is clearly insufficient.
- `main.go` should remain a thin executable entrypoint. Put runtime behavior in `app/` packages.
- Preserve package boundaries: config parses flags, source loads configured URLs, fetch downloads lists, proxy parses candidates, validate tests SOCKS5 reachability, output writes files, report defines schemas, pipeline orchestrates.
- Keep outputs deterministic. Sorted text output, stable report ordering, and stable CSV columns matter for comparing runs.
- Public proxy data is untrusted. Do not assume source metadata is truthful, and do not send sensitive traffic through discovered proxies.
- Use `context.Context`, explicit timeouts, and bounded concurrency for network work.
- For behavior changes, add or update tests in the nearest package. Existing tests are small and package-local; follow that style.
- Run `gofmt` on changed `.go` files. Do not format unrelated files.
- If changing source policy, read `docs/sources.md`. If changing output/report fields, read `docs/result-fields.md`. If changing package boundaries, read `docs/architecture.md`.

## Do not

- Do not commit generated artifacts from `generated/`, `dist/`, `bin/`, or ad hoc output directories.
- Do not treat local SOCKS5 reachability results as proof of CD quality. CD validation intentionally runs in GitHub Actions.
- Do not add release publishing behavior to the build workflow unless the user explicitly requests it.
- Do not broaden GitHub Actions permissions from `contents: read` without a concrete reason and user-facing explanation.
- Do not add scheduled or high-volume network validation jobs without explicit user approval.
- Do not silently change CLI flag aliases such as `-proxy`, `-concurrency`, or `-meta-output`; they are compatibility surface.
- Do not change CSV column names, JSON field names, or totals semantics without updating tests and the relevant docs.
- Do not hardcode real production credentials. Short-lived test fixtures are acceptable only when they are explicitly harmless and scoped to tests.

## Validation

Default validation after Go code changes:

1. `go test ./...`
2. `go vet ./...`
3. `go build -o bin/sky-socks5 .` when the CLI entrypoint, package exports, or build behavior changed.

Additional targeted checks:

- CLI flag or help changes: run `go run . -h` and the config tests through `go test ./app/config`.
- Parser/source changes: run `go test ./app/proxy ./app/source ./app/pipeline`.
- Output/report schema changes: run `go test ./app/output ./app/report ./app/pipeline` and compare `docs/result-fields.md`.
- Workflow changes: run `actionlint .github/workflows/*.yml` if available, plus `go test ./...` for CI-related edits.
- Default source-list changes: locally check fetchability, response format, and parser compatibility; do not claim local SOCKS5 reachability as CD evidence.
- Full proxy validation with `go run . -output-dir generated` requires external network and public proxies, writes artifacts, and may be slow or flaky due to upstream conditions.

For `AGENTS.md`-only changes, Go tests are not required unless you intentionally changed examples or command references. Confirm that only `AGENTS.md` files changed and that no `AGENTS.override.md` was created.

## Notes for future agents

- The runtime flow is documented in `docs/architecture.md`: `main.go -> app/pipeline -> app/config -> app/source -> app/fetch -> app/proxy -> app/validate -> app/output/app/report`.
- `docs/sources.md` is the policy source for default upstream list changes.
- `docs/result-fields.md` is the policy source for CSV and JSON report fields.
- `.github/workflows/cd.yml` owns live proxy validation and generated artifact upload; local runs are development smoke checks.
