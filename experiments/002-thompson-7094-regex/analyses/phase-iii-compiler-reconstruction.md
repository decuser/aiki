# Phase III — compiler reconstruction basis

## Source boundary

Thompson describes the compiler as three concurrently running stages:

1. a syntax sieve that admits syntactically valid regular expressions and inserts
   an explicit `.` operator for juxtaposition;
2. a conversion from the resulting infix expression to reverse Polish form;
3. an object-code producer accepting the reverse-Polish stream and emitting IBM
   7094 instructions.

Only Stage 3 is published in executable detail. Thompson explicitly calls the
first two stages straightforward and does not discuss their algorithms.
Accordingly, `compiler.ai` distinguishes provenance:

- `sieve` and `reverse_polish` are small conventional reconstructions satisfying
  Thompson's stated interfaces;
- `produce` reconstructs the published ALGOL-60 stage semantically, including
  its stack discipline and self-modifying object-code rewrites.

The initial reconstruction intentionally implements the compiler shown before
Thompson's later lambda/closure correction note. Consequently expressions such
as `a**` retain the pathological behavior Thompson discusses rather than being
silently repaired in the baseline compiler.

## Worked example

For `a(b|c)*d`, Stage 1 produces:

```text
a.(b|c)*.d
```

Stage 2 produces Thompson's reverse-Polish form:

```text
abc|*.d.
```

Stage 3 then emits 23 actual IBM 7094 words. The Phase-III corpus compares every
generated word against the previously reconstructed compiler-derived object
program.

## Location-4 source correction

The reconstructed Stage-3 compiler generates `TRA CODE+16` at location 4.
Phase II originally treated this as a correction to a printed `CODE+13`, but a
later page-image check established that Thompson's published listing itself is
`CODE+16`. The compiler therefore reproduces the published object program at
this location; the earlier discrepancy was in the project transcript.

## End-to-end criterion

The Phase-III gate is:

```text
regular expression
  -> reconstructed syntax sieve
  -> reconstructed RPN conversion
  -> Thompson Stage-3 code generation
  -> real 7094 instruction words
  -> Aiki 7094 emulator
  -> Thompson runtime
  -> matching signals
```

The generated object is loaded into a fresh emulated machine and exercised via
the same explicit `GETCHA` and `FOUND` host hooks used in Phase II.
