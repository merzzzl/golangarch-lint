# golangarch-lint

A self-contained architecture linter for Go. Deterministic, machine-readable, built for CI and AI-agent workflows.

This repository lints itself: [`.golangarch.yml`](.golangarch.yml) is a complete real-world config, the source tree is a living example of the conventions it enforces, and [`GOLANGARCH.md`](GOLANGARCH.md) is the instruction generated from it.

## Install

```bash
go install github.com/merzzzl/golangarch-lint/cmd/golangarch-lint@latest
```

## Usage

```bash
golangarch-lint lint [-config path] [-format text|json] [root]   # run checks
golangarch-lint docs [-config path] [root]                       # generate GOLANGARCH.md
```

Exit codes: `0` — clean, `1` — violations found, `2` — broken config or runtime error.

## Config

One `rules` array in `.golangarch.yml`. Every field is optional except `path`. `$module` expands to the module path from `go.mod`. Full reference by example — [this project's config](.golangarch.yml).

| Field | Meaning |
|-------|---------|
| `path` | glob over directories or file paths; file rules win over directory rules |
| `ignore` | per-rule path globs to skip |
| `mode` | directory shape: `any`, `flat` (no subdirs), `subdirs-only` (no files) |
| `allow-types` / `allow-vars` / `allow-funcs` | which declarations may exist: `all` (default), `local`, `exported`, `none` |
| `exclude-types` / `exclude-vars` / `exclude-funcs` | name globs exempt from all AST checks |
| `require-receiver` | functions must be methods: `all`, `local`, `exported`, `none` (default) |
| `exported.inputs` / `exported.outputs` | allowed types in exported signatures; stdlib always passes |
| `modules` | import whitelist, full paths; omitted = unrestricted, `[]` = stdlib only |

Type patterns in `exported`: exact `pkg.Name`, whole package `pkg`, one level `pkg/*`, any depth `pkg/**`.

Besides the rules, a few checks are always on: every function must live in a file named after it, package names must equal their directory names, and every directory with `.go` files must be covered by a rule.

Config validation fails fast (exit `2`): overlapping rule paths, unknown enum values, invalid globs.

## Docs for AI agents

`golangarch-lint docs` renders the config into [`GOLANGARCH.md`](GOLANGARCH.md) — a plain-language instruction an AI agent can follow when writing code in the repository. Point your agent instructions (e.g. `CLAUDE.md`) at it.
