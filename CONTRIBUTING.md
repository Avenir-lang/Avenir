# Contributing to Avenir

Thanks for contributing! This project is still in alpha; please open an issue
before major changes.

## Development Setup

Requirements:

- Go 1.22+

Build:

```bash
go build -o avenir ./cmd/avenir
```

Tests:

```bash
go test ./...
```

## Code Style

- Use `gofmt` for Go code.
- Keep error messages actionable and user‑facing.
- Avoid adding features in refactors.

## Pull Requests

Please include:

- A short summary of changes
- Tests added or updated
- Rationale for non‑obvious design choices
