# Reproducible experiment framework

Status: ACTIVE — hierarchy revision implemented and locally package-tested; authoritative `make validate` pending.

Prior related session: [`../2026-08-15-self-description-language-services/`](../2026-08-15-self-description-language-services/)

## Milestones

1. `01-experiment-generator.md` — ACTIVE; generator and experiment hierarchy implemented, authoritative gate pending.

## Current state

Aiki now has a proposed public `aiki experiment new <name>` scaffolding command. The running distribution's `experiments/` directory is the sole sequence authority; the generated experiment is created in the caller's current working directory for out-of-tree development.

The generated experiment home now separates procedure, observation, and analysis structurally:

```text
NNN-name/
  README.md
  experiment/
    PROCEDURE.md
    run.sh
  results/
  analyses/
```

Experimental source/materials live beside `run.sh` under `experiment/`. The runner uses `aiki` from `PATH`, records executable/version provenance with `aiki -v`, displays the complete transcript, and simultaneously writes it to a millisecond-timestamped file in `../results/`. Interpretations and later analyses belong under `analyses/`.

Top-level `experiments/README.md` defines this contract. `make dist` ships promoted experiments, and treecheck requires each distributed numbered experiment to retain the root README plus `experiment/PROCEDURE.md` and `experiment/run.sh`.

## Local evidence

The authoritative repository requires Go 1.24 while this environment contains Go 1.23.2 and cannot download toolchains. A disposable copy with only the `go` directive lowered to 1.23 was used for compile/test feedback:

```text
go test ./cmd/subcommands/tools/experiment ./cmd/subcommands/tools/treecheck  PASS
```

The generated-runner test uses a fake `aiki` that accepts only `-v`; it verifies that a run creates exactly one millisecond-timestamped transcript under `results/` and that the transcript contains the executable/version provenance.

## Exact next action

Merge the hierarchy revision and run authoritative `make validate`. Then, from an out-of-tree working directory, remove/rehome the earlier hand-built experiment and run:

```sh
aiki experiment new "Profiler calibration"
```

Confirm it creates `001-profiler-calibration/` with the new hierarchy. Rebuild Experiment 001 inside that generated home, run `experiment/run.sh`, retain the chosen reference observation under `results/`, and place its interpretation under `analyses/`.
