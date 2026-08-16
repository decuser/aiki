# Visible experiment demonstration

Status: ACTIVE

## Intent

Correct the mismatch between the experiment's strong end-to-end result and its nearly empty terminal presentation. Preserve the silent test corpora while making the historical reconstruction visible to a human running `./run.sh`.

## Finding

The runner already piped all output through `tee`; no output was being lost. The Phase I–III test modules intentionally emit nothing on success. The problem was therefore presentation architecture, not shell redirection.

## Change

Added a separate `demo.ai` program and appended it to `run.sh` after the regression corpora. It uses the reconstructed compiler to generate the object program, reports the compiler stages and word-4 historical discrepancy, loads the generated words into a fresh emulated 7094, and displays representative FOUND end offsets for matching and nonmatching inputs.

This keeps tests as assertions and makes the experiment itself the narrative demonstration.

## Validation

Source inspection complete. Local authoritative Aiki execution is required before GATED status because this change exists specifically to validate visible runtime output.

## Next action

Run `experiments/002-thompson-7094-regex/experiment/run.sh` and inspect the new demonstration section.
