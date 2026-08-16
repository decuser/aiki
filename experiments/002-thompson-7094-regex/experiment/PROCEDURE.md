# Procedure — Experiment 002: Thompson 7094 Regex Reconstruction

## Question

Can a deliberately minimal IBM 7094 written in Aiki reproduce the object code,
runtime behavior, and compile-search path described in Ken Thompson's 1968
regular-expression paper?

## Method

The reconstruction is staged so machine correctness is established before
historical runtime behavior, and historical runtime behavior before compiler
reconstruction.

```text
Phase I    machine
Phase II   published runtime and object program
Phase III  compiler and end-to-end reproduction
Phase IV   operator monitor and observability
```

The exact cut structure and machine scope are specified in
`../../../proposals/thompson-7094-regex.md`.

## Current procedure

Phases I and II are GATED under `aiki v0.4.0-alpha-26`. The current runner
exercises all three phases together:

```sh
./run.sh
```

It runs:

1. the complete Phase-I targeted 7094 machine corpus;
2. the Phase-II Thompson runtime/object corpus;
3. the Phase-III compiler and end-to-end corpus;
4. the Phase-IV monitor corpus; and
5. a scripted monitor demonstration.

Phase II installs Thompson's published `CNODE`, `NNODE`, `FAIL`, `XCHG`, and
`INIT` routines as 7094 instructions. `GETCHA` and `FOUND` are explicit host
boundaries because Thompson specifies the contract of the former and leaves the
latter application-dependent.

Phase III reconstructs the three compiler stages. Thompson states the contracts
of the syntax sieve and reverse-Polish stages but does not publish their
algorithms, so those two stages are explicitly reconstruction choices. The
object-code producer is reconstructed from Thompson's published ALGOL-60.

For the worked expression `a(b|c)*d`, the corpus requires:

```text
sieve:  a.(b|c)*.d
RPN:    abc|*.d.
object: 23 IBM 7094 words
```

Every generated object word is compared against the compiler-derived Phase-II
program, and the generated program is then loaded into a fresh emulator and
executed through Thompson's runtime.

The corpus uses Thompson's published `TRA CODE+16` at location 4. During Phase
IV the page image was rechecked against the project transcript and showed that
an earlier `CODE+13` reading was a transcription/OCR error, not a discrepancy in
Thompson's paper. The reconstructed Stage-3 compiler independently produces the
same `CODE+16` word.

Raw runner output is retained in `../results/`. Source reconciliation and
interpretation are retained separately under `../analyses/`.


## Operator monitor

The Phase-IV console is launched interactively with:

```sh
aiki console.ai
```

or with a command file:

```sh
aiki console.ai monitor-demo.cmd
```

Numeric monitor arguments are octal by default, matching IBM documentation. The
monitor owns no machine store directly: examine, deposit, register access, reset,
and stepping are performed through the spawned machine service. `input` restores
the prepared Thompson runtime/object program and sets the host character stream.
The monitor services `GETCHA`, EOF flushing, and `FOUND` explicitly so the
operator can step across the complete historical execution path.
