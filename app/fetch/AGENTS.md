# app/fetch navigation card

`app/fetch` downloads upstream proxy source lists and records fetch metadata for reports. Read this card before changing HTTP request behavior, proxy support, timeout handling, response parsing, or fetch result ordering. Key files: `fetch.go`; callers live in `app/pipeline`.

## Local invariants

- Fetches run concurrently but results are sorted by source URL before returning.
- The optional fetch proxy applies only to downloading source lists, not to SOCKS5 validation.
- Only 2xx upstream HTTP status codes are treated as successful fetches.
- Response bodies are parsed as trimmed non-empty lines with BOM prefixes removed.
- Fetch errors are captured in `Result.Error` so pipeline can continue reporting per-source failures.

## Local rules

- Preserve context-aware requests.
- Keep scanner limits intentional if response size handling changes.
- Prefer adding focused tests before changing subtle network behavior; avoid live upstream tests in unit tests.

## Do not

- Do not make one failed source abort the entire aggregate fetch unless the user explicitly changes that product behavior.
- Do not mix candidate parsing into this package; `app/proxy` owns proxy syntax.
- Do not use local source fetch success as evidence that proxies are reachable.

## Validation

Use root validation commands. If adding tests here, keep them local and use `httptest` instead of public network endpoints.
