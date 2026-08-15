# Relocatable release distribution

Status: ACTIVE — implementation complete; authoritative gates pending.

Proposal: `proposals/relocatable-release-distribution.md`

## Milestones

1. `01-contract-and-runtime-path.md` — executable location established as the distribution root.
2. `02-release-archive.md` — generated archive and distribution proof added.
3. `03-closeout.md` — documentation and final gate status.

## Current state

The runtime module registry now derives shipped module roots from the running
executable and does not recursively scan the process working directory for named
packages; local/path imports remain explicit. `make dist` generates a versioned unpacked distribution and matching archive as
siblings of the source tree, with the executable and `lib/` side by side.
`make distcheck` is the acceptance harness: it unpacks the archive into
a temporary prefix, changes to an unrelated working directory, invokes `aiki`
through `PATH`, and imports the shipped `list` module.

The local execution environment cannot download the Go 1.24 toolchain, so Go
compilation and the distribution harness have not been run here.

## Exact next action

On the authoritative development machine run:

```text
make distcheck
make validate
```

If both pass, mark this session and proposal COMPLETE, commit the project branch,
and integrate it with a non-fast-forward merge.
