# Milestone 29 — selfhost out-of-tree module resolution

Status: GATED

## Finding

Running the development `aiki` executable from `~/forge/dev/test` separated two
resolution defects that were previously hidden by repository-root CWD.

The first defect was incorrectly spelled sibling imports under `selfhost/`, such
as `./selfhost/runtime` from a file already inside `selfhost/`. Rewriting those
imports as genuine file-relative imports fixed one-level self-host execution
outside the distribution.

A second-level run still failed with `[@self_fault, :shape, "unknown shape:
error"]`. The remaining cause was different: the independent self-host loader
resolved *named* modules such as `selfhost/bootstrap` by probing `lib/...`
relative to process CWD. The host loader does not do that; it searches an
ordered module-root inventory derived from the executable distribution,
development working tree, and user library.

This distinction was confirmed by the out-of-tree smoke:

- one-level `bootstrap.run("1 + 2 * 3\\n", ...)` returned `9`;
- nested bootstrap failed before reaching the third-level expression.

The earlier profile counts for the nested case therefore represented an early
fault path and are not valid scale measurements.

## Correction

- Keep sibling imports genuinely file-relative (`./lexer`, `./runtime`, etc.).
- Add private HAL primitive `_module_roots()` returning a copy of the host
  registry's configured module search roots in lookup order.
- The blessed `lib/selfhost/bootstrap.ai` captures `_module_roots` and calls
  `evaluator.configure_module_roots(...)` before evaluation.
- The independent loader resolves named Aiki modules by walking those roots,
  while still performing its own file read, lex, normalize, parse, evaluation,
  export collection, and cache management.
- Preserve absolute paths in self-host path canonicalization; the earlier
  relative-only `clean_path` would otherwise strip the leading `/` from the
  host-provided roots.
- No public Aiki API is added. `_module_roots` is a platform/distribution fact
  available only through the blessed bootstrap boundary, analogous to the
  private HAL capabilities already captured there.
- The existing out-of-tree three-level self-interpretation invariant remains
  the authoritative regression gate.

## Gate

GATED. Authoritative `make validate` passed after the absolute-root harness correction. Direct execution from `~/forge/dev/test` then succeeded at one and two self-host levels; the cleaned profiling drivers both reached the expected final value `9`. The out-of-tree recursive profile was subsequently collected successfully.

## Validation follow-up

The first authoritative `make validate` after the named-module-root correction
failed three Phase-I authority invariants. The selfhost source imports were
correct; the tests were relocating `check_lexer_authority.ai`,
`check_normalization_authority.ai`, and `token_authority.ai` into `/tmp` before
execution. That destroyed the meaning of their now-correct sibling imports such
as `import("./grammar_authority")`. The older incorrect `./selfhost/...` spelling
had accidentally hidden the test defect through process-CWD fallback.

Correction: authority programs that depend on source-relative siblings are now
executed from their real repository paths via `runAikiFile`. Temporary-source
execution remains available for deliberately synthesized test harnesses. This
preserves source identity and makes the invariants agree with the out-of-tree
module-resolution rule rather than weakening it.

Local Go execution of the corrected invariant tests is environment-blocked by
the unavailable Go 1.24 toolchain; authoritative validation remains the gate.

## Follow-up — invariant root identity

Authoritative validation after the source-identity correction exposed a test-harness path bug: `distributionRoot()` returned the relative spelling `../..`. The authority tests then used that value both as the child process working directory and as the prefix of the source path, causing the child to resolve `../../selfhost/...` from an already-rooted working directory.

Correction: `distributionRoot()` now returns an absolute repository root. Source-relative programs can therefore execute from their real paths while the child process uses the repository root as its working directory, without double-applying a relative prefix. This is a harness correction only; self-host import semantics are unchanged.

Authoritative `make validate` subsequently passed; see Gate above.
