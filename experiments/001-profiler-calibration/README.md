# Experiment 001 — Profiler Calibration

This experiment establishes whether Aiki's semantic profiler can be trusted
before it is used to interpret the much larger measurements produced by
recursive self-hosting.

It starts with ordinary Aiki programs whose work can be counted directly from
the source, then carries the same repeated-work check through one and two
levels of self-interpretation.

## Layout

- `experiment/` contains the procedure, runner, and Aiki programs under test.
- `results/` contains raw transcripts produced by runs. The dated reference
  run retained with this experiment is `results/a promoted `reference-<timestamp>.txt``.
- `analyses/` contains interpretation of retained observations.

Start with:

```text
experiment/PROCEDURE.md
```

Then run:

```sh
cd experiment
./run.sh
```

Each execution is displayed on standard output and also recorded as a dated
file in `../results/`.

The interpretation of the retained reference run is in:

```text
analyses/interpretation.md
```


## Results and replication

`results/reference-2026-08-15.txt` is retained as the original reference
observation. Each later invocation of `experiment/run.sh` creates a separate
millisecond-timestamped `run-*.txt` result beside it.

The reference is not overwritten by later runs. Keeping both makes repeated
execution part of the evidence: stable work counts can be distinguished from
elapsed-time variation.

