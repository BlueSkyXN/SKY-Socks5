# Architecture

SKY-Socks5 is intentionally structured as a small Go CLI rather than a service or framework.

## Runtime flow

1. `main.go` passes command-line arguments into `app/pipeline`.
2. `app/config` parses and validates CLI flags.
3. `app/source` loads source URLs from `configs/sources.txt` or a custom file.
4. `app/fetch` downloads source lists concurrently.
5. `app/proxy` parses, normalizes, validates, deduplicates, and sorts proxy addresses.
6. `app/validate` tests each unique proxy through the configured probe URL.
7. `app/output` writes deterministic text and JSON outputs.
8. `app/report` defines the JSON report schema.

## Package boundaries

- `main.go` is the only executable entrypoint.
- `app/` contains application packages. It keeps the root tidy without splitting this small CLI into `cmd/` and `internal/`.
- `configs/` stores versioned runtime defaults such as source lists.
- `docs/` captures project decisions that should survive code rewrites.
- Generated artifacts belong in an output directory such as `generated/`, not in Git.

## Governance

Changes to `main` should go through pull requests with CI and CODEOWNERS review. Workflow files are part of the repository control plane and should be reviewed with the same care as application code.
