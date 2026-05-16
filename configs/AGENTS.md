# configs navigation card

`configs/` stores versioned runtime defaults, especially `configs/sources.txt`, the default aggregate upstream source list. Read this card before adding, removing, reordering, or replacing source URLs. Key files: `sources.txt`, `docs/sources.md`, `.github/workflows/cd.yml`.

## Why this is high-risk

- Default sources drive the CD proxy validation workload.
- Large or mixed-protocol feeds can inflate candidate counts without improving valid SOCKS5 output.
- Local network results can differ materially from GitHub Actions network results.

## Required before changes

- Read `docs/sources.md` for current source policy.
- Prefer SOCKS5-specific HTTP/HTTPS text endpoints.
- Keep additions small and reviewable.
- Locally check source fetchability, returned format, and parser compatibility when practical.
- Use CD per-source results, not local reachability, before promoting large or uncertain feeds into the default aggregate list.

## Do not

- Do not add meta-source URL lists that require nested-source support unless code support is added first.
- Do not add very large mixed-protocol feeds by default without evidence they contribute useful unique and valid SOCKS5 proxies.
- Do not keep sources that consistently return hard HTTP errors such as 404 or 502.
- Do not commit generated output from source-list experiments.

## Validation

- `go test ./app/source ./app/proxy ./app/pipeline`
- Full `go run . -output-dir generated` requires network and public proxies; treat it as a development smoke run, not CD proof.
