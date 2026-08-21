# Proposal: Immutable Number Literal Realization

## Status

**COMPLETE**

## Selection evidence

After removing speculative parser mismatch allocation, the remaining parser and
lexer costs are classified as startup. The strongest bounded execution-side
materialization shared by self-host and PDP is repeated exact number parsing.

Post-mismatch allocation evidence:

```text
three-level self-host
  NewNumberFromString
    flat        ~145.5 MB
    cumulative  ~363.0 MB  (~23.3%)

PDP-11 Cut 7 10x
  NewNumberFromString
    flat        ~13.0 MB
    cumulative  ~35.0 MB   (~12.6%)
```

The cumulative family includes `big.Rat` parsing/normalization machinery such as
`strings.NewReader`, `math/big.nat.make`, and exact-number normalization.

Code inspection establishes that this is repeated evaluator execution, not
source parsing:

```go
func (e *Evaluator) evalNumber(node *syntax.Node, env *value.Env) value.Value {
    num, err := value.NewNumberFromString(node.Value)
    ...
    return num
}
```

Every execution of the same immutable NUMBER AST node reparses the same source
text.

## Driver

> **A source literal is immutable. Its semantic value should not be reparsed on
> every observation.**

This is a realization/lifetime rule. Number semantics remain exact and
representation-private.

## Tranche A — evaluator-local source-number interning

Use a concurrency-safe evaluator-local cache:

```text
literal spelling -> immutable *Number
```

The key is literal text rather than AST identity.

Reasons:

- syntax remains independent of semantic/value;
- the cache does not retain obsolete ASTs in persistent evaluator/REPL use;
- equal literal spellings across loaded modules can share one realization;
- `sync.Map` permits spawned computations to evaluate through the same
  Evaluator safely.

First observation parses through the existing authoritative
`NewNumberFromString`. Later observations return the cached immutable Number.

Concurrent first observations may perform duplicate parsing, but `LoadOrStore`
selects one authoritative cached value. This affects only transient realization,
never value semantics.

## Scope boundary

Included:

- source `NUMBER` evaluation.

Excluded:

- `to_number(string)` and other dynamic conversion;
- arithmetic result caching;
- constant folding;
- syntax-node semantic fields;
- global/process-wide literal interning;
- string/rune/symbol caching;
- call-argument realization.

Each excluded family requires separate evidence.

## Safety requirements

The tranche survives only if:

- Number operations remain immutable;
- exact rational behavior is unchanged;
- the syntax package gains no dependency on semantic/value;
- shared-Evaluator concurrent observation is race-safe;
- persistent evaluators do not retain AST nodes through the cache;
- warm repeated literal observation allocates zero.

## Critical gate

Run:

```text
make validate
extra/profiling/alpha38-runtime-survey.sh
```

Success requires:

1. semantic counts unchanged;
2. `NewNumberFromString` and its big-number parsing family collapse from
   repeated-execution profiles;
3. self-host allocation and/or elapsed improve materially;
4. PDP remains favorable or materially flat;
5. Four-Way Life remains materially flat or improves;
6. no startup/parser redesign is introduced to rescue weak results.

If the cache merely replaces number parsing with comparable synchronization or
map overhead, revert it.


## Final gate — PASSED

The three-leg post-literal survey shows a broad execution-side win.

Three-level self-host:

```text
                         pre-literal       post-literal
elapsed                  8.415147775s      7.694529303s
alloc_bytes              1772281800       1342447120
mallocs                  47881922         29233795
gc_cycles                53               40
```

Approximate change:

- elapsed: -8.6%;
- allocated bytes: -24.3%;
- mallocs: -38.9%;
- GC cycles: -24.5%.

PDP-11 Cut 7 10x:

```text
                         pre-literal       post-literal
elapsed                  1.027890926s      0.928814488s
alloc_bytes              337069888        266185432
mallocs                  6845206          3656550
gc_cycles                9                8
```

Approximate change:

- elapsed: -9.6%;
- allocated bytes: -21.0%;
- mallocs: -46.6%.

Four-Way Life also improved across the coordinator and every worker. Coordinator:

```text
elapsed      584.932794ms -> 532.257041ms
alloc_bytes  73571200     -> 51886288
mallocs      2047707      -> 1076331
gc_cycles    5            -> 3
```

Workers A-D each lost roughly 2.5-3.2 MB and 109k-129k mallocs.

`NewNumberFromString` disappeared from the dominant self-host allocation
profile. Its small remaining presence in PDP is startup/dynamic conversion
rather than repeated source-literal realization.

## Completion decision

Completed 2026-08-21.

Evaluator-local literal-text interning survives. It changes semantic lifetime,
not numeric meaning:

> **A source literal is immutable. Its semantic value should not be reparsed on
> every observation.**

Dynamic `to_number(string)` remains uncached and outside this project.
