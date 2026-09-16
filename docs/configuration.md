# Configuration

[Documentation index](README.md)

The current format is version 2. Rules combine a scope with independent checks. Missing check blocks do not add restrictions, but directory coverage and package-name checks still apply.

```yaml
version: 2
rules:
  - scope:
      path: "**"
      ignore: ["**/*_generated.go"]
    layout:
      directories:
        flat: true
        subdirs: true
      binding:
        name: true
        receiver: true
    declarations:
      types: ["*"]
      vars:
        exported: []
        unexported: ["*"]
      consts: ["*"]
      funcs: ["*"]
    imports:
      allow:
        - "$module/internal/dto"
        - "$module/internal/services/**"
      deny:
        - "$module/internal/repos/**"
    signatures:
      exported:
        inputs: ["$module/internal/dto"]
        outputs: ["$module/internal/dto"]
```

This is a policy example, not a universal default. See the repository's [own config](../.golangarch.yml) for a complete multi-layer setup.

## Blocks

| Block | Purpose |
| --- | --- |
| [scope](scope.md) | Which files/directories the rule covers |
| [layout](layout.md) | Directory shape, allowed/required files and declaration placement |
| [declarations](declarations.md) | Allowed names by declaration kind and visibility |
| [imports](imports.md) | Imported packages and actual direct call targets |
| [signatures](signatures.md) | Receivers and parameter/result types |

`version` and each rule's `scope.path` must be specified. Unknown keys, invalid patterns, unsupported versions, multiple YAML documents and overlapping scope patterns are rejected. There is no separate `calls` block, `exclude` block or top-level `ignore`.

`$module` expands to the module path from `go.mod` when loading a project. Migration preserves the placeholder.

## Shortcut reference

| Short form | Equivalent full form |
| --- | --- |
| `vars: []` | `vars: {exported: {allow: []}, unexported: {allow: []}}` |
| `vars: ["Err*"]` | `vars: {exported: {allow: ["Err*"]}, unexported: {allow: ["Err*"]}}` |
| `exported: []` under a declaration kind | `exported: {allow: []}` |
| `unexported: ["*"]` under a declaration kind | `unexported: {allow: ["*"]}` |
| `imports: []` | `imports: {allow: []}` |
| `imports: ["$module/internal/dto"]` | `imports: {allow: ["$module/internal/dto"]}` |
| `binding: false` | `binding: {name: false, receiver: false}` |

Declaration shortcuts work for `types`, `vars`, `consts` and `funcs`. Use the full filter mapping whenever you need `deny`. These shortcuts do not apply to the `exported`/`unexported` groups in `signatures`.

## Absence is different from an empty list

- Missing `allow`, or an empty filter mapping `{}`, means unrestricted except for `deny`.
- Empty declaration allow lists forbid those declarations.
- `imports: []` allows standard-library imports and calls, plus calls within the current package.
- Empty signature `inputs`/`outputs` allow builtin and standard-library types; they do not mean zero parameters/results.
- `filenames: []` forbids checked Go files.
- `binding: {}` disables declaration-to-filename binding; `binding: false` permits at most one declaration per file.

Explicit `deny` takes priority over `allow` and dependency defaults. A declaration's permission does not exempt it from layout, imports or signature checks.
