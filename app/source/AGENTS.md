# app/source navigation card

`app/source` loads configured upstream source URLs from a local text file. Read this card before changing source-file parsing, empty-file behavior, BOM handling, or comment handling. Key files: `source.go`, `source_test.go`, `configs/sources.txt`.

## Local invariants

- Non-empty, non-comment lines are source URLs.
- Full-line `#` comments and blank lines are skipped.
- UTF-8 BOM prefixes should not break the first line.
- Empty effective source files must return an error.
- This package does not fetch URLs and does not validate whether a URL returns SOCKS5 candidates.

## Local rules

- Keep this package focused on local file parsing. Fetchability belongs to `app/fetch`; candidate parsing belongs to `app/proxy`.
- If accepted source-file syntax changes, update tests and check whether `docs/sources.md` needs a matching change.

## Do not

- Do not make source loading perform network requests.
- Do not silently accept an empty source set as a successful run.

## Validation

- `go test ./app/source`
- Run `go test ./...` before handoff when behavior affects pipeline startup.
