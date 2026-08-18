# Four-Way Life analyses

`experiment/profile.sh` writes semantic-count measurements here under
`analyses/profile/` by default.

The profiling split is deliberate. The coordinator and each worker are separate
Aiki OS processes, so profiling only `coordinator.ai` does not measure worker
Life scans. The harness therefore measures:

- coordinator at 1, 5, and 20 generations;
- each of the four worker domains for one deterministic generation;
- one correlated Go CPU profile for the five-generation coordinator run.

Use the counts to identify the dominant semantic work before changing the Life
algorithm. Profile individual workers with `--cpu` only after the count data
identifies which worker warrants deeper substrate correlation.
