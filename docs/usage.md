# Usage

[Documentation index](README.md)

## Installation

```sh
go install github.com/merzzzl/golangarch-lint/cmd/golangarch-lint@latest
```

From this repository, commands can also be run with `go run ./cmd/golangarch-lint` instead of `golangarch-lint`.

## Commands

```sh
golangarch-lint lint [-config path] [-format text|json] [root]
golangarch-lint docs [-config path] [root]
golangarch-lint migrate [-config path] [-output path] [root]
```

`root` defaults to the current directory and should contain the project's `go.mod`. Put flags before the positional root argument. Without `-config`, the tool looks for `.golangarch.yml`, then `.golangarch.yaml`, in that root. An explicit relative `-config` or `-output` path is relative to the working directory.

```sh
golangarch-lint lint .
golangarch-lint lint -config configs/architecture.yml -format json .
```

Go package loading uses the current build environment. Dependencies must be available; package compilation errors stop the analysis. Test files are excluded from architecture checks. Filesystem layout checks can still see non-test Go files excluded by build tags, while type and call analysis covers the loaded build.

## Results and CI

| Exit code | Meaning |
| --- | --- |
| `0` | No errors; warnings may be present |
| `1` | Architecture violations |
| `2` | Invalid arguments/configuration or loading/runtime failure |

```sh
golangarch-lint lint -format json . > architecture-report.json
```

A successful empty JSON report is:

```json
{"violations": [], "count": 0, "warning_count": 0}
```

Each violation has `severity`, `check`, `rule`, `path`, `pos` and `message`. `count` counts errors, not warnings. Examples of checks are `import-denied`, `call-denied`, `call-unresolved` and `layout-scattered`. Preserve the command's exit status in CI; warning-only reports succeed.

## Instructions for AI agents

```sh
golangarch-lint docs .
```

This creates or replaces `GOLANGARCH.md` in the project root. The file describes the effective configuration in natural language, including call restrictions and warning behavior. Both shorthand and full configuration forms produce equivalent instructions. Regenerate it after changing the config. The command renders instructions; it does not verify that project code complies.

For conversion commands, see [migration](migration.md).
