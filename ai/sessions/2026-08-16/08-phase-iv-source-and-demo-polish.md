# Phase IV — source correction and demonstration polish

Status: ACTIVE

## Intent

Close Phase IV with historically correct source presentation and a monitor
walkthrough that visibly exercises Thompson's list machinery rather than showing
only freshly-cleared list storage.

## Thompson source correction

The paper transcript was rechecked against the page image. Thompson's published
23-word `a(b|c)*d` listing has `TRA CODE+16` at object location 4. The earlier
project transcript that read `CODE+13` was wrong.

This supersedes the earlier Phase-II claim of a paper/compiler discrepancy.
The reconstructed Stage-3 compiler had generated `CODE+16` all along; it was
reproducing Thompson correctly.

Corrections made throughout the executable experiment and documentation:

- removed the false verbatim/corrected object-program fork;
- canonicalized the runtime API as `install_published_code` /
  `prepare_published`;
- changed Phase-II/III tests to assert published/compiler agreement;
- corrected the experiment README, procedure, analyses, and proposal;
- retained the old `CODE+13` finding only in provenance marked SUPERSEDED.

## Demonstration polish

The end-to-end demo now reports:

```text
published/compiler agreement
  published word 4: :TRA CODE+16
  generated word 4: :TRA CODE+16
```

A search with no FOUND signals is printed as `no match` rather than
`FOUND end offsets []`.

The scripted monitor walkthrough now:

1. displays registers and the first four generated instructions;
2. loads input `ad`;
3. traces six startup instructions;
4. uses asynchronous `run` and `wait` to complete the search;
5. displays retained CLIST and NLIST words after actual list activity;
6. displays the FOUND offsets;
7. resets and shows clean machine state.

This deliberately exercises both debugger-style stepping/tracing and the Cut-13
operator run-control surface.

## Build requirement

This cut changes only Aiki source, experiment documentation, and provenance.
**No Aiki rebuild is required.** The existing rebuilt executable is sufficient.

## Required gate

Run:

```sh
cd experiments/002-thompson-7094-regex/experiment
./run.sh
```

Required evidence:

- all four phase corpora pass;
- the end-to-end demo shows published/compiler `CODE+16` agreement;
- `axd` prints `no match`;
- the scripted monitor walkthrough runs to completion;
- CLIST/NLIST inspection shows the retained runtime list activity after the
  search rather than only the reset-time zero workspace.

If clean, mark this milestone GATED and Phase IV COMPLETE.
