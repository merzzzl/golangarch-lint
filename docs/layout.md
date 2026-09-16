# Layout

[Documentation index](README.md)

`layout` controls directory contents, filenames and placement of top-level declarations. Package names must also match their directory names; `main` and the project-root package are exempt from that name check.

## Directories

```yaml
layout:
  directories:
    flat: true
    subdirs: false
```

`flat` permits Go source files directly in the directory. `subdirs` permits subdirectories. They are independent and default to `true` when omitted.

| flat | subdirs | Structure |
| --- | --- | --- |
| `true` | `false` | Go files, no subdirectories |
| `false` | `true` | Subdirectories, no direct Go files |
| `true` | `true` | Both |
| `false` | `false` | Neither |

## Required and allowed filenames

```yaml
layout:
  filenames: [package.go, run.go]
```

Every checked Go file must match an item, and every item must match at least one file in each scoped directory. This example requires exactly those two Go files. Empty directories can produce missing-file errors. Non-Go files and test files do not count.

```yaml
layout:
  filenames:
    - package.go
    - "lint_*.go"
```

This requires `package.go` and at least one `lint_*.go`, allows multiple matching files and rejects other checked Go filenames. Patterns match basenames and cannot contain path separators. For example, `v[0-9]*.go` requires a digit after `v`, but permits arbitrary characters after that digit.

Omitting `filenames` disables this check. `filenames: []` forbids checked Go files. Ignored files do not satisfy required patterns. Constraints apply even to files with their own scopes.

## Binding modes

| Configuration | Meaning |
| --- | --- |
| Omitted or `binding: {}` | No declaration-to-filename check |
| `binding: false` | At most one top-level declaration per file |
| `binding: {name: true}` | Group by declaration name |
| `binding: {receiver: true}` | Group by the owning type |
| `binding: {name: true, receiver: true}` | Each whole file may satisfy either grouping mode |

An omitted partner flag does not enable another mode. `binding: true` is not supported. `false` is exactly equivalent to both flags being explicitly false:

```yaml
layout:
  binding:
    name: false
    receiver: false
```

### One declaration per file

`binding: false` counts top-level functions, methods, types, variables and constants. Each name in a grouped declaration counts separately. Imports, struct fields, interface methods and local declarations do not count. An empty file is allowed unless another check forbids it.

### Name binding

```yaml
layout:
  binding:
    name: true
    receiver: false
```

Exported declarations must match the normalized filename. Private names must start with that normalized name. Normalization ignores case, underscores and the `.go` suffix. This applies to methods, functions, types, variables and constants.

In `run.go`, `Run`, `runHelper` and `runDefault` satisfy name binding. An exported `Reset` does not. Declaration and signature permissions still apply separately.

### Receiver binding

```yaml
layout:
  binding:
    name: false
    receiver: true
```

A type `ExampleType` belongs in `example_type.go`; its methods belong with that type. Value, pointer and generic receivers are supported. Variables and constants must start with the owning type's normalized name, such as `ExampleTypeDefault` or `exampleTypeLimit`. Plain functions still follow name binding.

```go
// example_type.go
package example

type ExampleType struct{}

var exampleTypeDefault = 1

func (*ExampleType) Run() {}
```

When both flags are true, the entire file must pass at least one mode. Mixing modes declaration by declaration cannot bypass restrictions.

## Consolidation warnings

`layout-scattered` reports a coherent group unnecessarily spread over multiple files. For example, `Run` in `run.go` and `runHelper` in `run_helper.go` can be combined. Correctly named exported `Run` and `Reset` in different files are not such a group.

Warnings stay within the same effective rule and matching build constraints. Receiver-group suggestions require the type's correctly named declaration file in the package. Invalid type names or unrelated receivers remain errors. Warning-only reports exit successfully. Disabled binding and single-declaration mode do not suggest grouping.
