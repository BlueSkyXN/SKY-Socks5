# app navigation card

`app/` contains the Go packages that implement the CLI runtime. Read this card before changing any package under `app/`; then read a deeper package card when one exists. Key files: `config/config.go`, `pipeline/pipeline.go`, `proxy/proxy.go`, `validate/validate.go`.

## Local invariants

- Runtime flow stays explicit: config parses flags, source loads URLs, fetch downloads source lists, proxy parses candidates, validate probes SOCKS5 reachability, output writes artifacts, report defines schemas, pipeline orchestrates.
- Keep `main.go` thin. New runtime behavior belongs in an `app/` package, not in the executable entrypoint.
- Network operations must accept context, timeout, and bounded-concurrency controls.
- Outputs and reports must remain deterministic enough for repeated run comparison.
- Tests should stay package-local and table-driven where possible.

## Local rules

- Read the package-specific `AGENTS.md` before changing a package that has one.
- Use dependency injection already present in `app/pipeline` and `app/validate` instead of adding live network tests.
- When changing report or CSV fields, read `docs/result-fields.md` before editing code.
- When changing source parsing or default source expectations, read `docs/sources.md` before editing code.

## Do not

- Do not merge package responsibilities into a generic utility package without a clear caller-driven reason.
- Do not add global mutable state for run configuration or result accumulation.
- Do not make tests depend on live public proxies.

## Validation

Use root validation commands. For app-only code changes, `go test ./app/...` is the fastest focused check, followed by `go test ./...` before handoff.
