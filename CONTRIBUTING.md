# Contributing

Thank you for your interest in contributing to DocuGuard!

## Development Setup

1. Clone the repository
   ```bash
   git clone https://github.com/blueberrycongee/docuguard.git
   cd docuguard
   ```

2. Install dependencies
   ```bash
   go mod download
   ```

3. Build
   ```bash
   make build
   ```

## Code Style

- Run `go fmt ./...` before committing
- Run `golangci-lint run ./...` to check for issues

## Commit Convention

Use conventional commits:

```
type(scope): message
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

## Pull Request

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes
4. Push to the branch
5. Open a Pull Request

## Testing

Run tests before submitting:

```bash
go test ./...
```
