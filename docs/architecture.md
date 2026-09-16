# Internal architecture

[Documentation index](README.md)

The controller computes scope once, loads a shared project snapshot, and passes it to independent services. It holds concrete service pointers and calls them directly.

| Service | Responsibility |
| --- | --- |
| Scope | Rule selection, file overrides, ignores and directory coverage |
| Layout | Directory shape, package names, binding and consolidation warnings |
| Declarations | Names and visibility of top-level declarations |
| Imports | Source imports and direct call targets, including interface implementations |
| Signatures | Receiver requirements and parameter/result types |
| Templater | Natural-language instructions for `GOLANGARCH.md` |

Checking services receive only their own settings. They do not import one another or read one another's diagnostics. Scope selection and AST/type data are shared inputs. The imports service owns SSA and call-graph construction; there is no separate calls service. Dependency syntax is loaded only when an imports policy requires analysis.

The repository's [self-lint config](../.golangarch.yml) defines its own source organization:

- DTOs and configuration types have individual named files; configuration methods live with their receiver type.
- Package errors and entry functions (`Load`, `New`, `Migrate`) live in `package.go`.
- Each service contains `package.go` and `run.go`.
- Reusable glob, AST and Go type operations live in `helpers`, usable by config, controller and services.
- The controller exposes `Lint` and `Docs`; loading is performed inside `Lint`.
- The CLI entry file contains `main`. Historical configuration versions live in `internal/migrations`.

These are policies of this repository, not mandatory conventions imposed on every project using the linter.
