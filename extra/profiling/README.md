# Profiling sweep

These drivers exercise Aiki mechanisms using Aiki's own language and library
surface. They are intended to be rerun before and after implementation changes.
They are not cross-language benchmark claims.

Run from the repository root:

```sh
make build
extra/profiling/sweep.sh profile-results
```

The output directory also contains `manifest.txt` with the Aiki version, Go version, host, Git description, and generation time.

Each workload produces:

- `<name>.txt` — Aiki semantic counts, source attribution, and measured-interval
  Go realization totals;
- `<name>.cpu.pprof` — Go CPU profile with Aiki correlation labels;
- `<name>.cpu.top.txt` — a textual CPU top report;
- `<name>.cpu.tags.txt` — the Aiki correlation labels observed in the CPU profile;
- `<name>.allocs.pprof` — ordinary Go allocation profile;
- `<name>.allocs.top.txt` — a textual allocation top report.

CPU profile labels identify `aiki_layer`, `aiki_function`, `aiki_file`,
`aiki_line`, and `aiki_primitive`. Go's pprof labels are not attached to the
allocation profile, so allocation correlation is intentionally not claimed.
The measured-interval `alloc_bytes`, `mallocs`, and `gc_cycles` in the
text report are the paired realization totals for the Aiki evaluation interval.

Workloads:

- `01-evaluator-loop.ai` — evaluator arithmetic/comparison/while path;
- `02-tail-recursion.ai` — proper tail-call path;
- `03-list-append.ai` — persistent list growth and `_append` substrate work;
- `04-store-bits.ai` — module calls, mutable store, and bit operations;
- `05-regex-ffi.ai` — repeated Aiki-to-Go regex boundary crossing;
- `06-concurrency.ai` — synchronized spawn/channel/send/receive work.
