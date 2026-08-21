# Argument Store CPU Gate

This is a focused diagnostic for the active call-argument realization project.

It does not change Aiki semantics or runtime behavior. It records CPU profiles
for the three-level self-host and PDP-11 Cut 7 10x workloads, then renders flat,
cumulative, and argument-store-focused pprof reports.

Run:

```sh
extra/profiling/argument-store-cpu-gate.sh
```

Output stays outside the repository:

```text
/tmp/aiki-argument-store-cpu-gate/
/tmp/aiki-argument-store-cpu-gate.tar.gz
```

The gate exists to distinguish the CPU cost of reusable argument-frame
realization from unrelated evaluator work before making another representation
change.
