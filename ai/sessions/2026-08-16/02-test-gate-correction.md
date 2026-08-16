# Test gate correction

Status: ACTIVE

## Trigger

The first human-readable demonstration faulted in `compiler.reverse_polish` while the preceding Phase III corpus had appeared to pass. The fault was:

```text
compiler.ai:138:in 'reverse_polish': index out of bounds: -1
```

Retained failed demonstration transcript:

```text
experiments/002-thompson-7094-regex/results/run-2026-08-16-000654.651.txt
```

## Findings

Two independent defects were exposed.

1. Stage 2 used compound `and` conditions as if they provided control-flow short circuiting. The expression accessed `operators[length(operators)-1]` even when `operators` was empty. The parser now tests emptiness structurally before every top-of-stack access.

2. More importantly, `run.sh` invoked the assertion corpora as ordinary programs (`aiki *_test.ai`). The Aiki test builtins record assertion failures and faults, but the `aiki test` subcommand is what reports accumulated results and returns failure status. Consequently, the earlier Phase I–III runner executions proved that the files ran but did not constitute assertion gates.

## Correction

- `compiler.ai`: remove reliance on boolean short circuiting in both operator-stack scans.
- `run.sh`: execute all three corpora with `aiki test`.
- Preserve `demo.ai` as a separate human-readable end-to-end demonstration after all test gates pass.

## Provenance correction

Prior claims that Phases I–III were GATED/COMPLETE are SUPERSEDED as gate claims. Their implementation and retained run artifacts remain useful evidence, but the full reconstruction must be re-gated with the corrected test runner.

## Validation required

Run:

```sh
cd experiments/002-thompson-7094-regex/experiment
./run.sh
```

Required evidence:

- each phase prints a `PASS ...` line and test summary from `aiki test`;
- zero failed tests in all three corpora;
- the end-to-end demonstration runs to completion;
- the result transcript is retained.

Only then may Phases I–III and the reconstruction return to GATED/COMPLETE.
