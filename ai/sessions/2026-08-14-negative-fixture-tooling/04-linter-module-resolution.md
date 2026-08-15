# Cut 4 — linter module-resolution coupling

Status: **ACTIVE — authoritative full-tree gate pending**

## Discovery

The first full `make validate` after Cuts 1–3 reached valid behavior fixtures
that the old lint formatting-preflight short circuit had previously masked.
`break_before_smoke.ai` and `newline_block_smoke.ai` both use `use("list")`, but
lint reported `map` as undefined.

The fixtures are valid: runtime module loading resolves the public package name
`list` through the module registry to `lib/list/list.ai`, whose exports include
`map`. The linter instead maintained a separate filesystem heuristic that looked
for forms such as `lib/list.ai`. Thus the newly complete traversal exposed an
older linter/runtime resolution drift rather than a negative-fixture defect.

## Change

- Bare public module names used by `use(...)` now resolve through
  `substrate.ModuleRegistry`, using `DefaultModuleRoots`, matching runtime
  registry semantics including nested package layout and native-default aliases.
- Path imports remain distinct: names such as `./helpers` continue to resolve
  relative to the current source file (and then cwd), matching the runtime's
  path-vs-registry distinction.
- The checker caches its registry for the linted source rather than rescanning
  for every `use(...)` occurrence.
- A regression test places package `list` in a nested directory and verifies
  that `use("list")` binds exported `map`; the old filesystem heuristic cannot
  satisfy that test accidentally.

## Validation

Source formatting/static inspection completed. Package execution is environment
limited here because the container cannot fetch the Ebitengine dependency and
has Go 1.23 while the repository requires Go 1.24.

## Gate

Run the authoritative `make validate` in the user's normal environment. This is
the final project gate; no further cut is planned unless that gate exposes a
new defect.
