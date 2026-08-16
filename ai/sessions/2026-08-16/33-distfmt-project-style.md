# Distribution formatting and formatter-stable project style

Status: **ACTIVE — awaiting authoritative validation**

Baseline entering this cut: completed HAL redesign, final `make validate`, `make rigorous`, and visual smoke reported green by the user.

## Purpose

Add a project/distribution source-layout pass without turning canonical formatting into a configurable style system.

The resulting model is:

```text
go fmt / aiki fmt   canonical language formatting
aiki distfmt        project distribution/source presentation
```

`distfmt` selects expanded layouts for long, audit-heavy declarative structures. Canonical formatters must preserve those selected layouts afterward so the tools reach a fixed point rather than fighting each other.

## `aiki fmt` preservation rule

The canonical Aiki formatter remains non-width-based and does not decide which compact constructs should expand. It now preserves an explicitly multiline list, call argument list, or function parameter list and normalizes that expanded layout.

For lists and calls, expanded layout is determined from the immediate delimiter/element relationship: the first immediate element or argument begins on a later source line than the opening delimiter. This deliberately does **not** use the full descendant span, because a compact call may contain a multiline function body while its argument layout remains compact.

Compact source remains compact unless `distfmt` selects expansion.

## `aiki distfmt`

Added the `distfmt` tool command with the same operational surface as `aiki fmt`:

```text
-n   list files that would change; do not write
-p   print one restyled file; do not write
-b   create .bak before overwrite
```

Recursive traversal follows the formatter's safety conventions: hidden directories are skipped and declared parse-negative Aiki fixtures are not treated as normal source.

The command handles `.ai` and `.go` source. Current project-style rule uses a fixed 100-column threshold and expands selected long comma-delimited declarative forms. There is no user-configurable width or alternate style mode.

### Safety

`distfmt` canonicalizes/parses before restyling, reparses afterward, and refuses to write if the syntax tree changes.

- Go selection is AST-position-driven. Only actual elided composite literals used as map values are candidates; text inside strings/comments cannot be mistaken for Go structure.
- Aiki output is checked for structural AST equivalence and the `style -> aiki fmt` cycle is iterated to a bounded fixed point before writing.
- Writes use temp file + `fsync` + rename, mirroring `aiki fmt`.

## Make integration

The central `fmt` target is now:

```text
go fmt ./...
./aiki fmt ./...
./aiki distfmt ./...
```

All existing Make workflows that depend on `fmt` therefore finish in project/distribution style without duplicating the command in individual targets.

## Project style

Added `docs/style.md`. The governing rule is compact ordinary source, expanded long declarative/audit-heavy structures, using only formatter-stable forms.

The HAL authority policy is the motivating Go example and is laid out vertically for auditability while remaining `gofmt`-stable. The same policy is applied to selected long Aiki exports, tables, and nested declarative lists.

## Discoveries and corrections

Two real-tree probes corrected early assumptions before handoff:

1. **Descendant-span preservation was too broad.** A compact call such as `run("name", () { ... })` spans multiple source lines because of the function body, but its argument layout is not expanded. Preservation now examines immediate element/argument layout instead. A regression test protects this case.
2. **Line-oriented Go restyling was unsafe.** A map-looking line inside a raw string in `distfmt`'s own test file was initially selected. Structural verification correctly refused the write. Go restyling is now AST-position-driven, and a regression test protects raw-string contents.
3. **One-pass Aiki styling was insufficient for noncanonical input.** The first pass could canonicalize without reaching final project layout. `distfmt` now computes a bounded formatter/style fixed point before writing.

## Evidence obtained here

A disposable Go-1.23-compatible probe was used only because the container cannot download the repository-required Go 1.24 toolchain/Ebiten dependency.

Focused tests pass in that probe:

```text
go test ./engine/formatter ./cmd/subcommands/tools/distfmt
```

A copy of the complete post-HAL source tree was then exercised end-to-end:

```text
distfmt ./...
-> gofmt + aiki fmt ./...
-> distfmt -n ./...
changed=false
```

The cycle was repeated after regression-test additions and again reached `changed=false`.

The top-level canary/full command build cannot execute in this environment because uncached Ebiten cannot be downloaded. This is an environment limitation, not a reported compile/test failure.

## Gate

**ACTIVE.** Authoritative validation is required after merge because this cut adds a new CLI command, changes canonical Aiki formatter behavior, and inserts `distfmt` into the central Make formatting path.

Recommended gate:

```text
make validate
```

A clean gate should leave the tree unchanged on an immediate second `make fmt` / `aiki distfmt -n ./...` check.
