# Scope

[Documentation index](README.md)

```yaml
scope:
  path: "internal/services/*"
  ignore:
    - "**/*_generated.go"
    - "internal/services/cache/generated"
```

Paths are relative to the project root and use `/`. A directory scope applies to files directly in matching directories. Use a glob to cover descendant directories; every directory containing checked Go sources needs coverage. Use `.` for the root directory and `**` for all directories.

## Patterns

| Pattern | Matches |
| --- | --- |
| `internal/services/*` | One path segment after `services` |
| `internal/services/**` | Zero or more path segments after `services` |
| `internal/services/*/package.go` | A specific file in each immediate service directory |
| `internal/v?` | A single character after `v` |
| `internal/v[0-9]*` | A digit followed by any suffix |

`**` has recursive meaning when it occupies a complete path segment. Other segment patterns use Go-style glob matching, not regular expressions. Quote patterns beginning with `*` in YAML. Scope overlap detection is conservative: two wildcard patterns may be rejected as potentially overlapping even if their intended sets differ.

## Ignores

`scope.ignore` excludes a file or a matching directory and its descendants from that rule's checks. It is not a package-loader exclusion: ignored source can still be loaded for type/call analysis. Test files (`*_test.go`) are always skipped. The scope scanner also skips hidden entries.

Ignoring a path does not make an otherwise uncovered directory covered. Use a scope covering that directory and place the exclusion within it.

## File overrides

A file scope replaces the matching directory rule for declaration, binding, dependency and signature checks. It does not inherit omitted fields. Directory structure and filename-set constraints still apply.

```yaml
version: 2
rules:
  - scope:
      path: internal/example
    declarations:
      funcs: []
    layout:
      filenames: [package.go, example.go]
  - scope:
      path: internal/example/package.go
    declarations:
      types: []
      vars: []
      consts: []
      funcs:
        exported: [New]
        unexported: []
    imports: []
```

Here `New` is allowed only in `package.go`. That file must still belong to the directory's required filename set. If a file rule ignores the file, the directory rule is not used as a fallback.
