# Experiment framework — conceptual summary

Aiki experiments are being treated as distributed empirical investigations rather than tests. The crucial design distinction is between **procedure**, **observation**, and **interpretation**.

The experiment generator therefore does more than allocate a sequence number. It creates a small evidence-bearing home in which those roles are physically separated:

```text
NNN-name/
  README.md
  experiment/
    PROCEDURE.md
    run.sh
    materials...
  results/
  analyses/
```

The root README orients the reader. `experiment/` contains the preregistered-enough question, procedure, expected relationships, caveats, executable runner, and materials. `results/` contains what actually happened. `analyses/` contains what is inferred from those observations. This makes it harder to silently rewrite expectations after seeing a result and gives later readers a direct path from method to evidence to interpretation.

Experiment development remains out of tree. Sequence identity comes from the running Aiki distribution's `experiments/` collection, while creation occurs in the caller's current directory. Promotion into the distribution is manual and therefore curated.

The generated runner records the actual `aiki` executable and version, executes from the experiment directory, and uses `tee` so the human-visible transcript and retained result file are the same observation. Machine-dependent outputs are evidence, not correctness golds.

The first planned use is to recreate Experiment 001, profiler calibration, under this hierarchy rather than manually reshaping the earlier draft.
