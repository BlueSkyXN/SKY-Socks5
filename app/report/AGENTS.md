# app/report navigation card

`app/report` defines JSON report structures and validated proxy record fields. Read this card before changing JSON field names, totals, report shape, or validated proxy schema. Key files: `report.go`, `docs/result-fields.md`, `app/pipeline/pipeline.go`.

## Local invariants

- JSON field names are output contract, not internal-only implementation detail.
- `Metadata.Finalize` sets UTC timestamps, run duration, invalid proxy totals, and stable source ordering.
- `ValidatedProxy` must stay aligned with CSV row construction in `app/pipeline`.
- Full Cloudflare trace maps remain in JSON while the CSV stays compact.
- Per-source valid counts are restored after global deduplication, so repeated proxies count as valid for every source that listed them.

## Local rules

- Read `docs/result-fields.md` before schema changes.
- Update pipeline tests and docs when adding, removing, or renaming report fields.
- Keep optional JSON fields intentionally optional; avoid emitting noisy empty fields unless consumers need them.

## Do not

- Do not rename JSON tags casually.
- Do not add Cloudflare diagnostic fields to top-level CSV/report records without a documented consumer need.
- Do not change totals semantics without updating tests and workflow summaries.

## Validation

- `go test ./app/report ./app/pipeline`
- Review `.github/workflows/cd.yml` summary generation if report fields consumed by CD change.
