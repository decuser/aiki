### Chat bootstrap for the canvas redesign

Paste this at the top of a fresh chat.

Title
Aiki canvas redesign for repeatable Ebiten sessions

Goal
Make canvas sessions repeatable from the REPL even though Ebiten allows one window per process and cannot reopen after close. Do this by putting graphics behind a process boundary so the REPL can launch, close, and relaunch graphics sessions.

Canvas sessions spec

* Aiki provides interactive graphics via Ebiten.
* Ebiten supports one window per process and does not support reopening a window after close.
* Therefore, graphics must live behind a process boundary so the REPL can open and close graphics repeatedly.
* One executable supports multiple modes, for example `aiki --repl`, `aiki --canvas`. Default is text REPL mode.
* In REPL mode, graphics are ephemeral sessions: REPL launches canvas, canvas runs, canvas closes, control returns to REPL, sessions repeat.
* Unix strategy: Linux and macOS use fork so parent runs REPL, child runs Ebiten, child exits on close, parent continues.
* Windows strategy: spawn exec `aiki --canvas` for the same user experience.
* Initial interaction model: separate graphics process, console REPL stays pure. Single window combined UI is optional later.

Non negotiables

1. Single binary, multiple modes.
2. Default remains text REPL mode.
3. Repeatable canvas sessions must work without restarting the REPL.
4. Unix fork is primary path, Windows spawn exec is fallback with identical user flow.
5. Do not change language semantics, value types, parser, evaluator, fmt, lint, module system, or any non canvas subsystem.
6. No drive by refactors. Only touch code required for canvas sessions.

Scope
In scope

* CLI mode wiring for `--canvas` and any needed flags.
* Canvas lifecycle, child process management, liveness detection.
* Canvas handle semantics in the parent.
* Parent child message protocol for drawing primitives if required.

Out of scope

* Turtle library.
* New drawing primitives beyond what exists.
* Any redesign of errors, numbers, modules, concurrency, tooling.

Safety rails

* Keep changes isolated to the canvas subsystem plus minimal CLI wiring.
* After each phase, run `make validate`. It is the gate.
* If a change would touch code outside the canvas target, stop and ask first.

Work plan
Phase 0 baseline

* Tag current head.
* Identify current canvas entry points and any existing tests.
* Add a minimal repeatability check plan: open close open close in one REPL run.

Phase 1 process boundary

* Implement fork child canvas session on Unix.
* Implement spawn exec canvas session on Windows.
* Ensure parent survives repeated close cycles.

Phase 2 handle and protocol

* Define what a canvas handle is in the parent.
* Define how commands are delivered and what happens on close.
* Ensure failures return errors, never panics.

Phase 3 harden

* Validate cleanup: no leaked processes, no stuck pipes, no zombie children.
* Validate determinism: command ordering is stable.

How to work with me

* Ask early on ambiguity. Do not guess and run ahead. 
* Show code, not descriptions, when proposing changes. 
* Keep diffs small and scoped. After each phase, run `make validate` and report failures exactly.

