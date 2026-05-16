# .github/workflows navigation card

`.github/workflows/` contains GitHub Actions workflows for program tests, cross-platform build artifacts, and live proxy validation. Read this card after `.github/AGENTS.md` before editing any workflow YAML. Key files: `ci.yml`, `build.yml`, `cd.yml`.

## Local invariants

- `ci.yml` runs Go program tests only with `go test ./...`.
- `build.yml` creates Linux, macOS, and Windows archives as workflow artifacts; it does not publish GitHub Releases.
- `cd.yml` is manually dispatched proxy validation in the GitHub Actions network environment and uploads `generated/` artifacts.
- CD concurrency is `cd-proxy-validation` with `cancel-in-progress: true`.
- Workflow job summaries are bilingual and document scope, inputs, totals, and artifact fields.

## Local rules

- Keep `actions/setup-go` pointed at `go.mod` for the Go version.
- If report fields change, update the CD summary parser and `docs/result-fields.md` together.
- If source input behavior changes, update CD inputs and `docs/sources.md` together.
- Treat shell blocks as production automation: use explicit error handling and avoid silently ignoring failures.

## Do not

- Do not move live proxy validation into normal push/pull_request CI.
- Do not publish releases from `build.yml` unless the user asks for release automation.
- Do not broaden permissions beyond `contents: read` without a concrete workflow need.
- Do not remove CD artifact upload unless the result-delivery path is replaced.

## Validation

- `actionlint .github/workflows/*.yml` if installed.
- `go test ./...` when workflow changes affect CI commands.
- CD end-to-end validation must be confirmed in GitHub Actions; local `go run . -output-dir generated` is only a smoke check and requires network.
