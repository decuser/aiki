# Thompson 7094 regex reconstruction

Status: ACTIVE — Phases I–II GATED; Phase III implemented through Cuts 7–9 and awaiting authoritative runtime gate.

Prior session: [`../2026-08-15-experiment-framework/`](../2026-08-15-experiment-framework/README.md)

Proposal: `proposals/thompson-7094-regex.md`

Baseline: `v0.4.0-alpha-26-1-ga79cbc6` (`a79cbc6`).

## Phase plan

1. Phase I — Build the machine: Cuts 1–3 — GATED.
2. Phase II — Reproduce Thompson's machine-level result: Cuts 4–6 — GATED.
3. Phase III — Reconstruct the compiler: Cuts 7–9 — ACTIVE; implementation complete, runtime gate pending.

## Milestones

1. `01-proposal-baseline.md` — GATED.
2. `02-phase-i-cut-1-representation-state.md` — GATED.
3. `03-phase-i-instruction-execution-control.md` — GATED.
4. `04-phase-ii-runtime-object-reproduction.md` — GATED; combined Phase-I/II corpus passed under Aiki v0.4.0-alpha-26.
5. `05-phase-iii-compiler-end-to-end.md` — ACTIVE; three compiler stages and generated-object execution implemented; runtime gate pending.

## Current state

Phase I's targeted IBM 7094 emulator and Phase II's Thompson runtime/object
reproduction are runtime-gated under the repository Aiki executable.

Phase-II gate:

```text
Aiki version: aiki v0.4.0-alpha-26
result: experiments/002-thompson-7094-regex/results/run-2026-08-15-233403.309.txt
```

Phase III now reconstructs the syntax-sieve contract, RPN conversion, and
published ALGOL Stage-3 object-code producer. The generated object for
`a(b|c)*d` is required to match all 23 words of the compiler-derived Phase-II
program and independently generates `TRA CODE+16` at location 4.

The generated words are then loaded into a fresh emulated machine and executed
through Thompson's runtime. Source provenance distinctions are recorded in
`analyses/phase-iii-compiler-reconstruction.md`.

The new Phase-III Aiki files parse successfully under the current grammar. Full
execution remains environment-limited here because uncached Go dependencies
cannot be fetched.

## Exact next action

Run:

```sh
cd experiments/002-thompson-7094-regex/experiment
./run.sh
```

The runner executes all three phase corpora. If it passes, mark Phase III GATED
and the three-phase Thompson reconstruction COMPLETE. Any further work on
Thompson's lambda/closure correction or redundant-search optimization should be
a subsequent extension, not silently folded into the baseline reconstruction.
