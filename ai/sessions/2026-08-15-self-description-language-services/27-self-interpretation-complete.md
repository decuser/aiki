# Milestone 27 — Self-interpretation complete

Status: GATED locally — full nested proof succeeds; authoritative `make validate` remains the user-side gate for this final delta.

## Intent

Complete Cut III.6 without weakening the independent implementation. Measure the nested path, correct only evidenced bottlenecks or semantic mismatches, and preserve the proof that Aiki can run an Aiki-written interpreter which in turn runs Aiki source.

## Measurement

A bounded workload self-host-loading `selfhost/parser.ai` was profiled after the initial rune-snapshot correction. It took about 15.9 seconds, allocated about 4.1 GB cumulatively, performed about 58.5 million allocations, and recorded about 3.27 million Aiki calls. Source attribution showed the dominant work remained in lexer helper churn (`length`, `equal`, literal matching, lexeme construction), not parser recursion or environment lookup.

A second lexer refinement replaced repeated table scans with direct token dispatch, used early-return keyword checks (important because Aiki `and`/`or` do not short-circuit), and removed a second per-token position scan. The same bounded workload fell to about 10.9 seconds, 3.3 GB cumulative allocation, 39.3 million allocations, and 2.26 million Aiki calls. The reviewed lexical conformance fixtures remained exact.

## Discoveries and corrections

1. **String scanning realization cost.** Aiki string indexing materializes runes, so repeatedly indexing the entire source made the independent lexer effectively quadratic on large source files. `lib/string.chars` now uses a private linear HAL realization (`_chars`), while its public surface and semantics remain unchanged. The self-host lexer snapshots the source once and remains independently implemented.

2. **Path-import fallback.** The self-host module resolver initially tried only caller-relative resolution. The reference loader tries caller-relative first and then the current working/distribution directory. The missing fallback caused nested loading of `./selfhost/*` modules to return recoverable import errors. The self-host loader now implements the same two-step rule.

3. **Canonical module identity.** The same physical module could be cached under spellings such as `lib/selfhost/../../selfhost/lexer.ai` and `./selfhost/lexer.ai`, causing repeated nested loading of the interpreter stack. The self-host loader now canonicalizes `.` and `..` path segments in ordinary Aiki before existence checks/cache lookup. This mirrors the identity effect of Go's cleaned paths without sharing Go module machinery.

## Final proof

The durable invariant runs:

```text
Go interpreter
  -> Aiki-written selfhost/bootstrap
      -> self-host-loaded Aiki bootstrap/interpreter
          -> third-level source: 1 + 2 * 3
              -> 9
```

The result `9` is significant because it also confirms Aiki's specified left-to-right binary evaluation (`(1 + 2) * 3`) at the third level. In the local disposable runtime the complete proof finishes in roughly 30 seconds.

No host parsing, host evaluation of the target program, shared AST structures, or reflective module shortcut was introduced.

## Validation evidence

- reviewed Phase-I lexical fixtures remain exact after lexer optimization;
- the bounded parser-load profile improved as recorded above;
- the exact nested proof completed with output `9`;
- the proof is encoded as `TestSelfhostSelfInterpretation` with a 90-second ceiling;
- changed Aiki source is canonical-format parse-preserving under the extracted formatter.

## Exact next action

Run authoritative `make validate`. If it passes, mark Cut III.6 and Phase III GATED/COMPLETE and close the combined proposal implementation session.
