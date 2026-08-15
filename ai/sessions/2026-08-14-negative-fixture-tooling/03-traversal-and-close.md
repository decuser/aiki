# Cut 3 — Traversal hardening and closeout

Status: **ACTIVE — full authoritative gate pending**

## Result

- Formatter tests prove two undeclared malformed files are both reported.
- Linter formatting-preflight tests prove two undeclared malformed files are
  both reported.
- Formatter and linter tests prove declared parse-negative specimens are skipped
  while ordinary files remain in traversal.
- Smoke tests prove declaration and parser-failure observation cannot drift in
  either direction.
- Investigation found that smoke/enginesmoke/debug also return integer statuses
  that the top-level dispatcher discarded. All integer-returning subcommands
  now propagate their status via `os.Exit`, making validation failures effective
  at the shell rather than advisory.
- `buglist.md` marks B1 resolved and records the broader defects discovered.

## Final gate

Pending the user's single `make validate` on the authoritative Go 1.24 tree.
