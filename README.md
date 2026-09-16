# golangarch-lint

Architecture checks for Go: file layout, declarations, signatures, imports and actual call targets, including interface implementations.

## Quick start

```sh
go install github.com/merzzzl/golangarch-lint/cmd/golangarch-lint@latest
golangarch-lint lint .
```

Create `.golangarch.yml` in the project root:

```yaml
version: 2
rules:
  - scope:
      path: "**"
    imports: []
```

This starting policy permits standard-library dependencies and calls within the current package. Add your internal and external dependencies to `imports` before using it in an existing project.

## Documentation

- [Usage and CLI](docs/usage.md)
- [Configuration and shortcuts](docs/configuration.md)
- [Scope](docs/scope.md)
- [Layout](docs/layout.md)
- [Declarations](docs/declarations.md)
- [Imports and call analysis](docs/imports.md)
- [Signatures](docs/signatures.md)
- [Migration from v1](docs/migration.md)
- [Internal architecture](docs/architecture.md)

The project checks itself with [.golangarch.yml](.golangarch.yml). [GOLANGARCH.md](GOLANGARCH.md) contains generated natural-language instructions for AI agents.
