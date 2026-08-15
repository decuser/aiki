# 03 — Closeout

Status: PENDING GATE.

The README now presents the release archive as the normal Aiki user installation:
unpack it, add its directory to `PATH`, then use `aiki` for the REPL or scripts.
Building from source is a separate development path, where `./aiki` intentionally
selects the development build.

No package manager, installer, fixed prefix, or Aiki-specific environment
variable was introduced. Distribution output is kept outside the development
tree: `make dist` leaves the unpacked user distribution and its archive beside
the source directory.

## Pending authoritative evidence

```text
make distcheck
make validate
```

When both pass, update this milestone, the session README, summary, and proposal
status to COMPLETE/GATED without changing the implementation contract.
