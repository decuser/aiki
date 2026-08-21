# Proposal: Four-Way Life

## Status

COMPLETE — Gates 1–4, systems acceptance, and final showcase reconciliation are implemented and retained as Experiment 003.

## Purpose

Four-Way Life is the representative completion program for Aiki's portable-systems work. Four independent Aiki worker processes compute competing Life populations while a fifth Aiki coordinator owns the world, communication, rendering, and systems effects. The visible simulation is the demonstration; the five-process architecture underneath is the completeness proof.

## Architecture

Five actual Aiki OS processes participate. Workers never share coordinator memory. The coordinator alone owns the Store-backed authoritative grid and all effects. Workers receive immutable generation state over process stdin and return proposals over process stdout.

Generation N is immutable while all four workers compute N+1. The coordinator waits for all four completion frames, resolves/commits the next grid, renders it, and only then starts another generation.

## Protocol

Gate 1 uses newline-delimited structured text with exactly one message per line. The protocol is deliberately human-inspectable. Packed-byte framing is deferred unless later work provides a concrete reason for it.

Workers inherit the coordinator runtime environment through the normal process API unless a later gate deliberately modifies that environment. Store exists only in the coordinator process; workers never observe or share it.

## Determinism

Same seed plus the same external inputs must produce the same committed generation sequence. Wall-clock time may eventually choose a default seed, but an explicit seed controls acceptance runs.

## Terminal windows

The intended interactive presentation may arrange the five Aiki processes in separate host terminal windows plus the coordinator-owned canvas. Window creation is outside Aiki; the systems claim concerns the independent processes and their endpoints, not terminal-emulator automation.

## Gates

### Gate 1 — Deterministic multiprocess Life core

Build the coordinator, four actual worker processes, newline protocol, immutable-generation computation, barrier, deterministic Store commit, optional canvas rendering, and headless acceptance mode.

Success requires repeated headless runs with the same seed and fixed inputs to produce identical committed generation sequences.

### Gate 2 — Load-bearing worker domains

A exercises file/path/string/bytes; B environment/time/hash/random; C process/regex/number; D pure computational facilities. Each recurring activity materially affects proposals.

Worker C uses Aiki itself or a project-controlled helper rather than depending on an arbitrary host utility.

### Gate 3 — Systems spine

Add signal handling, TCP observers, terminal UI, file locking, and graceful process shutdown through the coordinator.

### Gate 4 — Hardening and acceptance

Add deterministic acceptance runs, headless CI path, protocol/failure-path tests, cleanup verification, documentation, and language-report framing.

## Success criterion

The showcase is complete when five independent Aiki processes produce one coherent deterministic simulation; workers communicate only through Aiki process/I/O facilities; the coordinator owns state/effects; load-bearing worker domains influence the simulation; the systems spine exercises the portable-systems surface; headless acceptance is automated; interactive execution is immediately intelligible; and `make validate` remains green.


## Implementation status — 2026-08-17

- Gate 1: accepted — deterministic five-process Life core and line protocol.
- Gate 2: accepted — load-bearing A/B/C/D worker domains with deterministic output.
- Gate 3: accepted — TCP observer, signals, terminal lifecycle, file locking/logging,
  and graceful shutdown.
- Gate 4: acceptance harness added — deterministic replay, deliberate Worker C
  subprocess failure, systems-spine acceptance, and repository validation.

The worker helper-failure injection is deliberate: `FWL_HELPER_FAIL=1` causes
the project-controlled helper to exit 7. Worker C must treat the failure as
local and the simulation must continue to its requested generation limit.

The language-report framing remains:

> Four independent Aiki worker processes compute competing Life populations
> while a fifth Aiki coordinator owns the world, communication, rendering,
> and systems effects. The visible simulation is the demonstration; the
> five-process architecture underneath is the completeness proof.


## Final showcase reconciliation — 2026-08-17

The interactive presentation now matches the intended five-process framing.

- The coordinator runs in the launch terminal and owns authoritative state,
  canvas rendering, systems I/O, and four worker process handles.
- Four additional terminal windows display live status emitted by the actual
  A/B/C/D workers.
- Those windows are views only; worker protocol stdout remains private to the
  coordinator, preserving the validated process-endpoint architecture.
- The host terminal emulator is launcher convenience and is not part of the
  Aiki portability claim.
- TCP observer service remains independently connectable through `observer.ai`.
- Closing the canvas is a normal lifecycle event: generation scheduling stops,
  workers and resources are shut down, and the worker-view terminals close
  with the coordinator.

The showcase is therefore both visually faithful to the original concept and
architecturally faithful to the portable-systems completeness work.
