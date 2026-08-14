# Milestone 13 - spawned fault propagation

## Problem

A spawned computation could fault before sending an expected message. The fault was printed to stderr and discarded, leaving another computation blocked forever in `recv` or `select`.

## Decision

Do not add join, task handles, futures, channel close, or hidden completion messages.

A spawned fault is reported to the runtime. Blocking `send`, `recv`, and `select` observe the runtime fault signal and propagate the fault instead of waiting forever. Recoverable `[@error, ...]` values remain ordinary Aiki values and may be sent through channels normally.

This is intentionally narrower than general asynchronous exceptions: it fixes abandoned blocking communication without adding a task abstraction or polling every AST node.

## Implementation

- added optional `hal.AsyncFaultSource` runtime capability;
- `GoRuntime` retains the first pending spawned fault in a buffered channel;
- `EvalContext` exposes that signal to HAL primitives;
- `spawn` reports terminal faults to the runtime rather than swallowing them;
- blocking `send` and `recv` select between their channel operation and a pending spawned fault;
- evaluator `select` includes an internal runtime-fault arm that is not visible in Aiki syntax and is not counted as an Aiki receive.

## Validation

Race-detector tests pass for:

- spawned fault interrupting `recv`;
- spawned fault interrupting `select`;
- existing select behavior;
- existing spawn behavior;
- profiling/counter concurrency paths.

## Restart

Next ordered item: mundane systems substrate, split into args/env, directory listing, and random-access file operations.
