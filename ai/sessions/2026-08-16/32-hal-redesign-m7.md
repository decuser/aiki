# HAL redesign — M7 compatibility removal and final hardening

Status: **GATED**

Baseline entering M7: M6 passed `go test ./...` on the user's authoritative tree.

## Compatibility removal

Removed the obsolete per-command Canvas runtime bindings:

```text
_dot _line _rect _fill_rect _circle _fill_circle
_arc _clear _set_bg _set_fg _pen_size _set_turtle
```

They are no longer registered, captured by the selfhost bootstrap, or present in trusted-source dependency policy. Production Canvas/turtle code was already migrated to `_canvas_command` in M6. The runtime binding count falls from 132 to 120; canonical host contracts remain 38. An executable registry invariant requires the obsolete names to remain absent.

The Go helper functions that implement individual Canvas protocol operations remain private substrate implementation helpers used only by `halCanvasCommand`; they are not runtime bindings or public contracts.

## Canonical host authority

M4 separated authority from lexical `Scope`, but host grants were still represented by raw runtime binding names. M7 completes the Phase-II/III design: host authority is now keyed by canonical HAL identity.

Example:

```text
Aiki implementation binding:  _file_open
Authority required:            HAL.file.open
Substrate provenance:          go:os.Open/os.Create/os.OpenFile
```

`GoRuntime.authorityKey` translates a registered canonical host binding to its `HostOperation.Authority`. Non-HAL primitives (language/value primitives, providers, services) continue to use implementation-name grants. `AuthorityForSource` records actual raw source dependencies but translates host dependencies into canonical grants.

Executable evidence proves that `HAL.file.open` authorizes `_file_open` while raw `_file_open` authority does not. The three-name invariant also verifies every trusted Aiki host binding receives its canonical HAL identity and not its raw host binding as authority.

## Final runtime-ownership correction

The M7 package-global sweep found one M3 omission: Aiki test-framework result state (`passed`, `failed`, failure list, current test, test file) still lived in package globals. It is now `GoRuntime.testState`. Test service primitives are bound methods on the owning runtime.

The test CLI now creates one `GoRuntime` per test file and uses new caller-owned runner entry points (`RunWithRuntime`, `RunWithCountersRuntime`) so the command reads test results from the same runtime that evaluated the Aiki test. Existing `runner.Run` / `RunWithCounters` behavior remains unchanged; they continue to construct and own their runtime. A runtime-isolation test proves test results do not leak between runtimes.

## Documentation and coupling cleanup

- Added `docs/hal.md` as the durable non-session architecture description.
- Corrected stale comments equating `ScopePrelude` with HAL access.
- Updated runtime/authority comments to distinguish canonical host identities from non-HAL primitive grants.
- Strengthened executable contract coverage for canonical host authority.
- `gofmt` and `git diff --check` are clean.

## Local validation limitation

This environment still has Go 1.23.2 while the repository requires Go 1.24.0, and network access is unavailable for toolchain download. Local `go test` therefore cannot execute. No failing repository test was observed here; execution is blocked before compilation by the unavailable toolchain.

## Final gate

**GATED.** The user reported the final `make validate` passed on the authoritative tree after two contained validation corrections:

- stale boundary/contract tests were updated to assert canonical host authority (`HAL.io.print`) rather than the superseded `ScopePrelude`-implies-builtin-access rule;
- pure-Aiki `path.normalize` was corrected for Aiki's eager, left-to-right evaluation semantics by grouping arithmetic and replacing an unsafe boolean guard with explicit structural control flow.

The final gate includes build, Go formatting, Aiki formatting/lint/treecheck, `go test ./...`, and `./aiki test ./...`. M7 and the HAL redesign implementation are complete.
