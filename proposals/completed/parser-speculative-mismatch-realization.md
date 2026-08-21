# Proposal: Parser Speculative Mismatch Realization

## Status

**COMPLETE**

## Selection evidence

The post-parser-failure alpha-38 survey added Four-Way Life as a third
independent evidence leg.

Allocation-space evidence:

```text
PDP-11 10x
  Lexer.Tokenize       68.1 MB
  parseTerminal        57.5 MB
  parseProduction      40.5 MB

Life coordinator
  Lexer.Tokenize       10.8 MB
  parseTerminal         8.5 MB
  parseProduction       5.0 MB
```

All four Life workers are similarly dominated by parser/lexer startup.
Allocation-object evidence is even stronger: `parseTerminal`, `fmt.errorf`, and
`errors.New` account for a large fraction of object creation in PDP and every
worker process.

The completed parser-failure project removed speculative diagnostic-history
materialization. It did not remove speculative mismatch error objects.

## Problem

The recursive-descent parser currently uses freshly constructed Go `error`
values as ordinary control flow for grammar mismatch.

For terminals it also eagerly constructs a quoted expectation string:

```text
fmt.Sprintf("'%s'", term.Value)
```

on every mismatch, even though nearly all such mismatches are discarded while
alternatives are explored.

Neither object is semantic parser output.

## Driver

> **A speculative terminal mismatch is parser control flow. It should not
> allocate an error object.**

Diagnostic meaning remains authoritative. Speculative transport is negotiable.

## Tranche A

Preserve the grammar, AST, parser backtracking rules, best-failure selection,
and final rendered `SourceError`.

Change only mismatch realization:

1. use one shared zero-allocation internal no-match signal;
2. record raw terminal expectation in parser failure state;
3. quote/materialize that expectation only in final failure rendering;
4. keep structural parser errors (`undefined production`, unknown grammar
   expression) as real errors;
5. require repeated terminal mismatch itself to allocate zero.

This tranche deliberately does not alter lexer token representation,
parse-node allocation, grammar analysis, or AST construction.

## Critical gate

Run:

```text
make validate
extra/profiling/alpha38-runtime-survey.sh
```

Survival requires:

- exact parser diagnostics and whole-tree validation remain green;
- `parseTerminal`, `fmt.errorf`, and `errors.New` allocation objects collapse
  materially in PDP and Four-Way Life;
- self-host does not regress;
- full Five-Way Life behavior/timing remains valid.

Only after this gate may tokenization or parser node realization be considered.


## Final gate — PASSED

Post-tranche three-level self-host:

```text
elapsed      8.415147775s
alloc_bytes  1772281800
mallocs      47881922
gc_cycles    53
```

Compared with the immediately preceding three-leg survey:

```text
alloc_bytes  1791356352 -> 1772281800
mallocs      49249375   -> 47881922
```

Self-host elapsed was materially flat; allocation frequency fell.

PDP-11 Cut 7 10x changed much more strongly:

```text
                         pre              post
elapsed                  1.198536927s     1.027890926s
alloc_bytes              431394288       337069888
mallocs                  13652326        6845206
gc_cycles                13              9
```

The `fmt.errorf` / `errors.New` speculative mismatch cluster disappeared from
the dominant allocation-object profile.

Four-Way Life workers also improved materially:

```text
worker A  26.39 MB / 853066 mallocs -> 18.99 MB / 323498
worker B  26.35 MB / 853378 mallocs -> 19.03 MB / 323813
worker C  26.39 MB / 853213 mallocs -> 19.06 MB / 323639
worker D  27.65 MB / 894661 mallocs -> 20.47 MB / 365092
```

## Completion decision

Completed 2026-08-21.

The parser still allocates to tokenize source and construct an AST, but those
remaining costs are predominantly one-shot startup work. This project therefore
stops at removal of pathological speculative control-flow allocation rather
than redesigning legitimate parse-result carriers.

Surviving driver:

> **A speculative terminal mismatch is parser control flow. It should not
> allocate an error object.**
