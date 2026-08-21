# Experiment 002 — Thompson 7094 Regex Reconstruction

This experiment reconstructs Ken Thompson's 1968 regular-expression search
implementation by building the required subset of an IBM 7094 in Aiki,
reproducing the published object-level example, and then reconstructing the
compiler.

The governing design and phase gates are recorded in:

```text
../../proposals/completed/thompson-7094-regex.md
```

The experiment proceeds in three phases:

```text
I     build the machine
II    reproduce Thompson's machine-level result
III   reconstruct the compiler
```

Raw observations belong in `results/`; interpretation belongs in `analyses/`.

## Historical-source handling

The Phase-II source transcription was rechecked against the paper image during
Phase IV. Thompson's published object listing uses `TRA CODE+16` at location 4,
exactly as the reconstructed Stage-3 compiler generates. An earlier project
transcript had misread that field as `CODE+13`; that conclusion is retained only
in provenance as a superseded transcription/OCR error. The executable corpus
contains the published `CODE+16` form.
