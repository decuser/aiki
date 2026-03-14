### Chat bootstrap for canvas v2 refactor

Paste this at the top of a fresh chat.

Title
Aiki canvas v2: reenvisioned canvas primitives and mechanics for fast Logo style graphics

Goal
Replace the current canvas subsystem with a new canvas v2 that preserves the same top level user experience (create a canvas, draw immediately, close and reopen repeatedly), but is fast by design under heavy draw loops and supports a better set of drawing primitives. Canvas v2 remains backed by Ebiten in a child process so the REPL can open and close windows repeatedly.

Repo snapshot
I am working from this repo zip: `/mnt/data/aiki.zip`.

This bootstrap is modeled after:
- canvas session bootstrap
- canvas transport bootstrap
- lexer refactor bootstrap

Primary user goal
Immediacy like GW BASIC and Logo: user creates a canvas, issues a drawing call, sees results quickly, repeats without ceremony. Under heavy loops the REPL stays responsive and the window does not stall.

Non negotiables

1. Single binary, multiple modes.
2. Default remains text REPL mode.
3. Repeatable canvas sessions must work without restarting the REPL.
4. Unix fork is primary path, Windows spawn exec is fallback with identical user flow.
5. No drive by refactors. Only touch code required for canvas v2 plus minimal CLI wiring.
6. Do not change language semantics, value types, parser, evaluator, fmt, lint, module system, or any non canvas subsystem.
7. Failures return errors, never panics.
8. Deterministic ordering: the visual result for a given program is stable for a given backend and quantization policy.

Canvas v2 design targets

A. API contract
- Pixel coordinates by default.
- Aiki numbers are rationals; canvas converts to pixel coordinates using a declared rounding rule (round to nearest, tie rule defined).
- Background and border behavior are not exposed as a concept to the user.
- Canvas handle remains a value the user can pass around.

B. Performance contract
- Must remain responsive under heavy draw loops (spirograph scale).
- Minimize per op overhead by batching and by using binary framing.
- Child must be able to redraw correctly on expose events without requiring the parent to replay commands.

C. Primitive contract
Minimum primitives for v2
- dot
- line
- polyline
- rect and fill_rect
- circle and fill_circle
- arc
- clear
- destroy
Plus state setters
- set_fg
- set_bg
- pen_size

Optional primitives for later, not in v2 unless trivial
- ellipse
- bezier curve
- text
- path stroke and fill

Scope
In scope
- New canvas v2 protocol and child renderer implementation.
- Parent side canvas v2 handle, session management glue, and encoding path.
- Rounding policy implemented once at the parent boundary (or as fixed point mapping, but it must be declared).
- Tests for protocol, lifecycle, determinism, and performance safety.
- Minimal prelude surface changes if required to keep user facing names stable.

Out of scope
- Turtle library.
- Redesign of the language, numbers, errors, modules, concurrency, tooling.
- Any unrelated refactor of existing engine packages.
- Flags for selecting canvas versions. Canvas v2 becomes the only canvas once landed.

Safety rails
- Keep changes isolated to the canvas subsystem plus minimal CLI wiring.
- After each phase, run `make validate`. It is the gate.
- If a change would touch code outside the canvas target, stop and ask first.
- Keep diffs small, phase scoped, and reversible.

Work plan

Phase 0 baseline and inventory
- Tag current head.
- Identify current canvas entry points:
  - Parent builtins and value handle type
  - Session spawn and reap logic
  - Protocol encode decode and child dispatch
  - Existing tests and smoke programs
- Run `make validate` and record baseline.
- Add a quick repeatability scenario: open close open close in one REPL run.

Phase 1 spec and skeleton for v2
- Write a short canvas v2 spec block:
  - primitive list and arguments
  - rounding rule and tie rule
  - color encoding contract
  - pen size semantics
- Create new v2 packages or files under the canvas subsystem:
  - protocol constants: version and opcode ids
  - typed command structs per opcode
  - encoder and decoder
  - child dispatch switch
- Add a DecoderObserver hook that logs decoded commands in a debug friendly format (optional, off by default).

Phase 2 child renderer as retained display list
- In the child, keep a retained list of drawing ops.
- On each Ebiten draw, clear background then replay the list.
- Ensure clear resets the list and background state correctly.
- Ensure destroy closes the window and exits cleanly.

Phase 3 parent side bridge and batching
- Replace the parent command writer with v2 binary framed records.
- Batch writes so heavy loops produce fewer syscalls.
- Ensure close and teardown still reaps processes correctly.
- Ensure errors are returned to the evaluator layer without panic.

Phase 4 primitive completeness
- Implement polyline and arc in the child.
  - Arc can be flattened to segments internally, using a fixed step rule.
- Ensure circle remains a true circle under the chosen coordinate contract.

Phase 5 harden and tests
Protocol tests
- Golden byte vectors for each opcode.
- Decode exact bytes and compare structs.
- Round trip: cmd equals decode(encode(cmd)).

Lifecycle tests
- Open close open close in one test run.
- No leaked processes, no stuck pipes, no zombies.

Determinism tests
- For a fixed program, decoded command stream is stable.
- Observer output is deterministic.

Performance smoke
- Run the existing spirograph style program and verify it completes in a reasonable time and the window remains responsive.

Acceptance tests

1. `make validate` passes.
2. Canvas sessions repeat in one REPL run: open close open close.
3. No flags required to run canvas.
4. Under heavy draw loops, the parent does not stall on per op overhead (batching is effective).
5. No panics on malformed frames, unknown opcodes, or unsupported versions, only structured errors.
6. User code that previously drew basic primitives still works or has a small, documented update with a prelude shim.

How to work with me
- Ask early on ambiguity. Do not guess and run ahead.
- Show code, not descriptions, when proposing changes.
- After each phase, run `make validate` and paste the exact failures.
