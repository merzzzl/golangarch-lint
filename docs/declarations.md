# Declarations

[Documentation index](README.md)

`declarations` filters top-level names, separated by declaration kind and visibility.

| Kind | Declarations |
| --- | --- |
| `types` | Types, including interfaces and type aliases |
| `vars` | Package-level variables |
| `consts` | Package-level constants |
| `funcs` | Functions and methods |

Local variables/constants, struct fields and interface method entries are not top-level declarations. Exported/unexported follows Go visibility rules.

## Full form

```yaml
declarations:
  funcs:
    exported:
      allow: [Run, "Read*"]
      deny: [ReadUnsafe]
    unexported:
      allow: ["*"]
```

Patterns match names, not package paths. `deny` wins even if `allow` matches. An omitted allow list is unrestricted; `[]` allows no names. An omitted kind or visibility group is unrestricted.

## Equivalent shortcuts

These three fragments are equivalent:

```yaml
declarations:
  vars:
    exported:
      allow: []
    unexported:
      allow: []
```

```yaml
declarations:
  vars:
    exported: []
    unexported: []
```

```yaml
declarations:
  vars: []
```

A list at the kind level applies to both visibility groups. This also works with nonempty lists and with `types`, `consts` and `funcs`:

```yaml
declarations:
  types: ["*"]
  vars: ["Err*"]
  consts: []
  funcs:
    exported: [Run]
    unexported: ["*"]
```

`vars: ["Err*"]` gives both groups the same pattern; Go naming rules mean that uppercase `Err*` naturally matches exported names. To allow all private variables, write `unexported: ["*"]` explicitly instead.

## Package entry file

```yaml
scope:
  path: internal/example/package.go
layout:
  binding: {}
declarations:
  types:
    exported: [Service]
    unexported: []
  vars:
    exported: ["Err*"]
    unexported: []
  consts: []
  funcs:
    exported: [New]
    unexported: []
signatures:
  exported:
    receiver: false
```

Allowing `New` does not exempt it from signatures, imports or layout. This rule explicitly disables binding so the type, constructor and errors can share a file. See [scope overrides](scope.md).
