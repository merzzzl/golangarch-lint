# Migration from v1

[Documentation index](README.md)

```sh
golangarch-lint migrate -config .golangarch.yml -output .golangarch.v2.yml .
golangarch-lint lint -config .golangarch.v2.yml .
```

Without `-output`, converted YAML is written to stdout. With `-output`, a new file is created; existing files are never overwritten. Do not redirect stdout to the source config. `$module` placeholders are preserved. Review the converted file and warnings before replacing the original.

A v1 config can also be read by `lint` or `docs` through in-memory migration. This leaves the source file untouched and prints migration notes to stderr.

## Field mapping

| v1 | v2 |
| --- | --- |
| `path`, `ignore` | `scope.path`, `scope.ignore` |
| Global `ignore` | Copied into each rule's `scope.ignore` |
| `mode` | `layout.directories.flat` and `subdirs` |
| `file-binding` | `layout.binding.name` and `receiver` |
| `modules` | `imports.allow` |
| Declaration modes | `declarations` name filters by visibility |
| `require-receiver` | `signatures` receiver settings |
| `exported` signature types | `signatures.exported` |

Old variable permissions apply to both `vars` and `consts`. Migration writes full filter mappings, which have the same meaning as the [shortcuts](configuration.md).

## Semantic changes

Old exclusions could skip layout and signature checks in addition to declaration permissions. V2 declaration allow lists do not skip other checks. Conversion preserves declaration permissions and reports these differences; constructor files may need explicit rules.

Binding now checks types, variables and constants as well as functions/methods. The unified `imports` policy also restricts direct implementation calls, including calls through interfaces. V1 did not check those targets, so conversion emits a note to review dependencies.

An earlier v2 draft with a separate `calls` key is not accepted by the current parser or migration command. Consolidate its policy into `imports` explicitly: unioning different allow lists may weaken restrictions, while intersecting them may reject previously allowed behavior.

The legacy schema and conversion live in [internal/migrations/v1.go](../internal/migrations/v1.go); the current schema lives in [internal/config](../internal/config).
