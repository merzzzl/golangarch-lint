# Imports and call targets

[Documentation index](README.md)

`imports` is the single dependency policy. The imports service checks both source imports and the possible implementation packages at direct call sites.

```yaml
imports:
  allow:
    - "$module/internal/dto"
    - "$module/internal/services/*"
  deny:
    - "$module/internal/repos/**"
```

Patterns match package import paths, not function names or filesystem paths. `$module` expands from `go.mod`. `*` matches one path segment; `**` matches any depth.

## Shortcuts and defaults

```yaml
imports: []
```

is equivalent to:

```yaml
imports:
  allow: []
```

Nonempty lists work the same way:

```yaml
imports:
  - "$module/internal/dto"
  - golang.org/x/tools/go/packages
```

- Missing `imports`, `{}`, or omitted `allow` allows all dependencies except explicit `deny` matches.
- Standard-library imports and calls are allowed by default, even with `[]`.
- Calls inside the current package are allowed by default; other packages in the same module are not.
- `deny` overrides every default, including the current package and standard library.
- Builtins such as `len` are not package calls.

The standard-library classification uses the first import-path segment: it is treated as standard library when it contains no dot. Use normal domain-qualified module paths to avoid classifying your module as standard library.

## Interfaces and function values

```text
usecase -> repos                     checked as a direct dependency
usecase -> interface -> repos        checked against the implementation package
usecase -> service -> repos           each immediate dependency checked separately
```

The last chain is allowed when `usecase` permits `service` and `service` permits `repos`. The analysis does not forbid a transitive dependency merely because the first layer cannot call it directly.

A caller may import only a contract while invoking a repository implementation. That implementation must also be permitted by `imports`. Allowing it permits a direct source import too: there are no separate import and call permissions.

Likewise, permitting a package for its types, constants or blank import permits direct calls to it. The old `calls` configuration key is rejected rather than silently merged.

## Diagnostics and limits

- `import-denied`: a forbidden package is imported.
- `call-denied`: a possible direct target belongs to a forbidden package.
- `call-unresolved`: analysis could not determine a target; emitted as a warning.

One operation can produce both import and call violations. SSA and variable type analysis propagate receiver and function values; all discovered possible targets must be allowed. Analysis runs when an imports filter has an explicit allow list or deny patterns. Reflection, unsafe and runtime plugins are not fully modeled. Results cover the current Go build environment, not every platform or tag combination.
