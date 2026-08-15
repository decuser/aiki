# Aiki experiments

Experiments are reproducible empirical investigations distributed with Aiki.
They are not correctness gates and are not run by `make validate`.

Each numbered experiment separates four things deliberately:

```text
NNN-name/
  README.md                 orientation
  experiment/
    PROCEDURE.md            question, method, expectations, caveats
    run.sh                  executable procedure
    ...                     experimental materials
  results/                  raw observations produced by runs
  analyses/                 interpretation and subsequent analysis
```

This separation is part of the experiment contract: procedure is stated before
observation, observations are retained without reinterpretation, and analyses
are kept distinct from both.

New experiments are normally developed out of tree:

```sh
cd /path/to/work

aiki experiment new "Experiment name"
cd NNN-experiment-name/experiment
./run.sh
```

The running Aiki distribution's `experiments/` directory is the numbering
authority. The new directory is created in the caller's current working
directory. When the experiment is ready for distribution, promote the complete
numbered directory into this `experiments/` directory manually.

The generated runner uses `aiki` from `PATH`, prints the resolved executable
and version, displays the experiment transcript, and simultaneously writes the
same transcript to a dated file under `../results/`.

Machine-dependent measurements are observations, not behavioral invariants.
A result file may be retained as a dated reference observation when it is part
of the evidence for an analysis.
