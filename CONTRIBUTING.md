# Contributing

We welcome contributions! Please follow these guidelines:

## Development Setup

```bash
git clone https://github.com/vishu-commits7/distributed-raft-storage.git
cd distributed-raft-storage
go mod download
make protoc
```

## Code Style

- Follow standard Go conventions
- Run `gofmt` before submitting
- Add tests for new features
- Keep commits atomic and well-described

## Testing

```bash
make test
```

All tests must pass before a PR is accepted.

## Submitting a PR

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Commit changes: `git commit -am 'Add feature'`
4. Push to branch: `git push origin feature/my-feature`
5. Open a Pull Request

## Reporting Issues

Please include:
- Steps to reproduce
- Expected vs actual behavior
- Go version and OS
- Relevant logs or stack traces

## Architecture

See [README.md](README.md) for architecture overview.

## Questions?

Open an issue or discussion for questions about the codebase.
