# 01 — experiment generator

Status: ACTIVE — hierarchy revision implemented; authoritative gate pending.

## Intent

Formalize creation of reproducible Aiki lab experiments without turning experiments into correctness tests or forcing experiment development into the source tree.

## Sequence and creation authority

`aiki experiment new <name>` has two deliberately separate authorities:

- **sequence authority:** the `experiments/` directory beside the real running `aiki` executable;
- **creation destination:** the caller's current working directory.

The executable path is resolved through symlinks so a development symlink still discovers the experiments belonging to the actual Aiki distribution. Numbering scans only directories of the form `NNN-*`, takes the maximum existing number plus one, never fills gaps, and rejects duplicate sequence numbers or exhaustion at 999. Destination-local numbered directories do not influence the sequence.

## Experiment hierarchy

The original two-file scaffold was superseded during design discussion by a structure that keeps procedure, observations, and interpretation distinct:

```text
NNN-name/
  README.md                 orientation
  experiment/
    PROCEDURE.md            question, rationale, method, expectations, caveats
    run.sh                  executable procedure
    ...                     experiment-specific materials
  results/                  raw run transcripts / observations
  analyses/                 interpretations and subsequent analyses
```

There is intentionally only one README at the experiment root. `PROCEDURE.md` is not another orientation document; it records the method to be executed. Analyses are not embedded in results.

`run.sh` is executable, runs from `experiment/`, requires `aiki` on `PATH`, prints the resolved executable and `aiki -v`, and sends the full run through `tee` to a millisecond-timestamped file under `../results/`. This keeps the displayed run and retained observation identical.

## Distribution integration

`experiments/README.md` defines the hierarchy and experiment contract. Finished experiments are promoted there manually. `make dist` copies the full experiments collection into a release. `treecheck` recognizes artifacts under correctly numbered experiment directories and requires each distributed experiment to retain:

```text
README.md
experiment/PROCEDURE.md
experiment/run.sh
```

`results/` and `analyses/` may contain arbitrary intentional experiment artifacts; they are evidence and analysis, not correctness golds.

## Local evidence

The repository requires Go 1.24 but the local environment provides Go 1.23.2 and cannot download the required toolchain. In a disposable copy whose only compatibility edit was lowering the `go` directive to 1.23:

```text
go test ./cmd/subcommands/tools/experiment ./cmd/subcommands/tools/treecheck  PASS
```

The experiment package tests now verify:

- numbering comes from the distribution, not the destination;
- the hierarchy and executable runner are created;
- duplicate distribution sequence numbers are rejected;
- symlinked executable discovery reaches the real distribution;
- the generated runner invokes `aiki -v` and writes a millisecond-timestamped transcript under `results/`.

## Gate

ACTIVE pending authoritative `make validate` and one direct out-of-tree invocation with the real development `aiki`.
