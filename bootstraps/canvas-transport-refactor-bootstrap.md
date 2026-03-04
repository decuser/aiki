### Chat bootstrap for the canvas binary protocol refactor

Paste this at the top of a fresh chat.

Title
Aiki canvas protocol refactor to binary framed messages with decoder observer

Goal
Replace the current line delimited JSON canvas message protocol with a binary protocol to reduce per op overhead and improve responsiveness under heavy draw loops. Keep the user facing canvas API unchanged. Preserve repeatable canvas sessions from the REPL.

Repo snapshot
I am working from this repo zip: `/mnt/data/aiki.zip` (you will cut and provide the repo).
This bootstrap is modeled after the prior canvas redesign bootstrap  and the lexer refactor bootstrap .

Non negotiables

1. Single binary, multiple modes.
2. Default remains text REPL mode.
3. Repeatable canvas sessions must work without restarting the REPL.
4. Do not change language semantics, value types, parser, evaluator, fmt, lint, module system, or any non canvas subsystem.
5. No drive by refactors. Only touch code required for canvas messaging and session glue.
6. No flag first rollout. Land binary as the only protocol in this change.
7. Keep a decoderObserver that emits JSON for debugging, derived from decoded command structs, not from the wire bytes.
8. Command ordering must remain stable and deterministic for a given program.
9. Failures return errors, never panics.

Scope
In scope

* Parent child message protocol replacement: JSON lines to binary framed records.
* Parent bridge writer and child stdin reader updates.
* Decoder observer hook and optional logging output.
* Protocol tests: encode decode vectors and round trip.
* Minimal CLI wiring only if required for observer enablement (for example env var), but protocol itself is not selectable.

Out of scope

* Turtle library.
* New drawing primitives beyond what exists.
* Any redesign of errors, numbers, modules, concurrency, tooling.
* Any changes to Ebiten rendering semantics.
* Any refactor of canvas session lifecycle not required by the protocol change.

Protocol requirements

* Framing: length prefixed records.

  * u32 little endian payload length
  * payload: u8 version, u8 opcode, then opcode specific fields
* Versioning:

  * Reject unsupported versions with an error shape.
* Stable opcodes:

  * Explicit numeric ids, never auto numbered.
  * Do not rely on file order or init order for ids.
* Field encoding:

  * Keep it boring and fixed width at first unless you already have varint utilities.
  * Integers as i32 or i64 where needed.
  * Colors as packed RGBA u32.
  * Strings as u32 length plus bytes.
* DecoderObserver:

  * Called after successful decode into a typed command struct.
  * Can render JSON lines for debug and can be disabled by default.

Work plan

Phase 0 baseline

* Tag current head.
* Identify current protocol entry points:

  * Parent command encode and pipe writer path.
  * Child stdin read loop and decode path.
  * Any protocol tests that exist.
* Run `make validate` and record baseline.

Phase 1 protocol skeleton

* Create a canvas protocol module inside the canvas subsystem.
* Define:

  * Protocol version constant.
  * Opcode const block with explicit ids.
  * One command struct per op, field order matches wire order.
  * Encoder and decoder functions per op.
  * Decode and dispatch switch in the child.
  * DecoderObserver interface and hook.

Phase 2 wire swap

* Replace JSON encode in parent with binary encode.
* Replace line scanner JSON decode in child with framed binary reader and decoder.
* Preserve existing canvas API calls in Aiki and Go wrapper functions.

Phase 3 harden and tests

* Golden vectors:

  * For each opcode, encode known struct and compare exact bytes.
  * Decode exact bytes and compare struct.
* Round trip:

  * cmd equals decode(encode(cmd)) for each opcode.
* Stress safety:

  * Ensure child read loop handles partial reads and multiple frames per read.
  * Ensure close and teardown still reaps processes correctly.
* Gate: run `make validate` after each phase and report failures exactly.

Acceptance tests

1. All existing canvas and integration tests pass under `make validate`.
2. No flags are required to run canvas.
3. DecoderObserver can be enabled to emit JSON lines and matches the decoded command stream deterministically.
4. Canvas sessions remain repeatable in one REPL run: open, close, open, close.
5. No panics on malformed frames, unknown opcodes, or unsupported versions, only structured errors.

How to work with me

* Keep diffs small and scoped.
* Show code, not descriptions, when proposing changes.
* Do not edit unrelated files or run formatting across unrelated directories.
* After each phase, run `make validate` and paste the exact failures.

