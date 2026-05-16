# .github navigation card

`.github/` is the repository control plane for ownership and automation. Read this card before changing `CODEOWNERS`, workflow permissions, triggers, branch-gate assumptions, or automation policy. Key files: `CODEOWNERS`, `workflows/ci.yml`, `workflows/build.yml`, `workflows/cd.yml`.

## Why this is high-risk

- Workflow changes can alter CI, build artifacts, and live proxy validation.
- Permission changes affect the repository security boundary.
- `CODEOWNERS` changes affect review routing for protected paths.

## Required before changes

- Read the specific workflow file you are changing.
- For workflow edits, also read `.github/workflows/AGENTS.md`.
- Keep default workflow permissions at `contents: read` unless the user asks for broader behavior and the change explains why.
- Preserve CODEOWNERS coverage for `.github/` unless the user explicitly changes ownership policy.

## Do not

- Do not add release publishing, write permissions, or token use without explicit user approval.
- Do not remove manual `workflow_dispatch` CD controls without a clear replacement.
- Do not hide network-heavy validation in ordinary CI.

## Validation

- `actionlint .github/workflows/*.yml` if `actionlint` is installed.
- `go test ./...` for CI-related changes.
- CD behavior requires GitHub Actions network conditions and cannot be fully proven by local runs.
