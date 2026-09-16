# Signatures

[Documentation index](README.md)

`signatures` restricts receivers and the types of parameters/results, separately for exported and private functions and methods.

```yaml
signatures:
  exported:
    receiver: true
    inputs: ["$module/internal/dto"]
    outputs: ["$module/internal/dto"]
  unexported:
    receiver: true
    inputs:
      - "$module/internal/config"
      - "$module/internal/dto"
    outputs: ["$module/internal/dto"]
```

## Receiver

| Value | Meaning |
| --- | --- |
| Omitted | Functions and methods allowed |
| `true` | A receiver is required |
| `false` | A receiver is forbidden |

This checks receiver presence, not its concrete type. File ownership belongs to [layout binding](layout.md). Use a separate constructor file rule when constructors need `receiver: false` but other declarations require methods.

## Parameters and results

`inputs` and `outputs` are allow lists. An omitted list is unrestricted. `[]` allows builtin and standard-library types; it does not require an empty signature.

Patterns may be an exact qualified type, a package, or a package glob:

```yaml
signatures:
  exported:
    inputs:
      - "$module/internal/dto.Request"
      - "$module/internal/model"
      - "$module/pkg/**"
    outputs: []
```

Here `string`, `bool` and `error` results are allowed. A project-defined result type needs a matching entry. Types from the current package are not implicitly exempt.

Pointers, slices and arrays are unwrapped before the type check. The current implementation treats maps, channels and function types as builtin containers; their nested types are not recursively restricted. A generic type's package matching uses the outer type's package, not package names inside its type arguments.

## Plain helper functions

```yaml
signatures:
  exported:
    receiver: false
    inputs: []
    outputs: []
```

The `exported: []` shortcut from declarations does not apply here. Signature visibility groups must be mappings, and `inputs`/`outputs` have no `deny` form. A permitted signature does not override declaration-name or dependency restrictions.
