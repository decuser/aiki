# Proposal: Speculative Parser Failure Realization

## Status

**ACTIVE**

## Baseline

Release baseline: `v0.4.0-alpha-38`, commit `36d5a0e`.

Authoritative self-host baseline:

```text
elapsed      8.792440682s
alloc_bytes  1928120376
mallocs      49812106
gc_cycles    64
```

Fresh alpha-38 pprof survey (sampling totals differ slightly from semantic
counter totals by design) selected parser failure bookkeeping:

```text
selfhost alloc_space
  Parser.recordFailure   191.56 MB   10.43%

PDP Cut 7 10x alloc_space
  Parser.recordFailure   617.67 MB   60.87%
```

It is the only remaining allocation site that is simultaneously material in
self-host and dominant in the PDP systems workload.

## Problem

The recursive-descent/backtracking parser records the best speculative failure
seen so far. The alpha-38 implementation materializes more diagnostic state
than the final error actually needs:

```go
stack := make([]string, len(p.stack))
copy(stack, p.stack)
p.failure = &ParseFailure{..., Stack: stack}
```

Every failure at or beyond the furthest token therefore:

1. copies the full active production stack;
2. constructs a replacement heap `ParseFailure`;
3. maintains the parser production stack solely so that the copy can later be
   searched for one fact.

`renderFailure` uses that copied stack for exactly one purpose: find the
**innermost active production carrying grammar `@error` metadata**.

The full stack is therefore realization history, not required diagnostic
meaning.

## Semantic boundary

Parser acceptance, backtracking, furthest-failure selection, exact source
position, `got`, `expected`, newline-termination provenance, grammar error
metadata selection, and final rendered diagnostics are authoritative.

The speculative representation used to remember those facts is not.

Working driver:

> **The parse failure is diagnostic meaning. Speculation need not materialize
> diagnostic history.**

## Design

### 1. Maintain the relevant error production incrementally

`Parser` holds one scalar `errorProduction` containing the innermost currently
active production whose metadata has `@error` text.

When entering a production:

- save the previous scalar;
- if the new production has `@error`, replace it;
- restore the previous scalar on exit.

This is exactly equivalent to scanning the old production stack from innermost
to outermost during final rendering.

The parser production stack is otherwise unused and is removed from normal
parser realization.

### 2. Store best failure in-place

Replace:

```text
failure *ParseFailure
```

with:

```text
failure    ParseFailure
hasFailure bool
```

`recordFailure` overwrites the stable in-parser record rather than allocating a
new object for every speculative replacement.

### 3. Preserve compatibility

`ParseFailure.Stack` remains in the exported struct for source compatibility,
but normal parser operation leaves it nil. `ParseFailure.Production` carries
the relevant production directly.

`renderFailure` prefers `Production` and retains a legacy Stack fallback for
manually constructed `ParseFailure` values.

## Invariants

- furthest-token comparison remains unchanged;
- equal-position replacement (`>=` behavior) remains unchanged;
- every accepted program remains accepted;
- every rejected program retains its final failure location and diagnostic;
- closing-delimiter raw errors still override production metadata;
- grammar newline-continuation diagnostics remain unchanged;
- `recordFailure` allocates zero during speculative replacement;
- normal parser operation never materializes `ParseFailure.Stack`.

## Gate

One critical gate only:

1. `make validate` PASS;
2. self-host semantic/call realization counts unchanged;
3. PDP semantic/call realization counts unchanged;
4. `Parser.recordFailure` disappears or becomes negligible in allocation-space
   and allocation-object profiles;
5. self-host allocation/elapsed favorable or materially flat;
6. PDP allocation drops materially, with elapsed favorable or materially flat.

If those conditions pass, close the proposal. Do not add parser memoization,
packrat tables, AST arenas, token interning, or grammar redesign in this wave.
