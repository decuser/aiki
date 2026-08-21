# Proposal: Immutable String Observation Realization

## Status

**ACTIVE**

## Baseline

Release baseline: `v0.4.0-alpha-36`.

Authoritative three-level self-host baseline after Adaptive Persistent List:

```text
elapsed      10.619609147s
alloc_bytes  8901681936
mallocs      50401888
gc_cycles    267
```

Runtime-realization survey repeat:

```text
elapsed      10.804215561s
alloc_bytes  8901480296
mallocs      50401663
gc_cycles    267
```

The repeat is consistent with the alpha-36 baseline.

## Survey evidence

Allocation-space attribution for the three-level self-host:

```text
evalStringIndex                           6171.30 MB   73.20%
NewCallEnv                                 931.20 MB   11.04%
evalCallArgs                               210.01 MB    2.49%
Parser.recordFailure                       192.56 MB    2.28%
NewNumberFromString                        159.51 MB    1.89%
Env.Set                                    156.04 MB    1.85%
```

One evaluator site therefore owns 73.2% of all allocated bytes.

Inspection shows that `evalStringIndex` performs:

```go
runes := []rune(s.Val)
return runes[i]
```

for every logical string index operation.

The operation needs one rune. The implementation realizes every rune.

This is an observation-materialization pathology, not yet evidence that Aiki
strings need multiple physical representations.

## Governing semantic principle

Aiki strings are immutable values.

> **The string is immutable. Observation need not materialize it.**

This proposal deliberately does **not** yet adopt an adaptive string
representation.

The broader project principle remains:

> Semantic properties are authoritative. Physical representation is
> negotiable.

But alternate string representation is admissible only if the post-fix evidence
shows that flat UTF-8 itself remains the limiting realization.

## Scope

Included:

- central rune-aware string observation authority;
- allocation-free rune indexing;
- allocation-free rune count;
- allocation-free first-rune observation;
- allocation-free rune-wise comparison;
- migration of evaluator string indexing;
- migration of observational `first`, `length`, `ord`, and string comparison;
- Unicode/invalid-UTF-8 equivalence tests;
- zero-allocation test for `RuneAt`;
- focused witness;
- self-host and PDP evidence;
- whole-tree validation.

Not included:

- cached rune arrays;
- rune-offset indexes;
- ropes;
- substring objects;
- small-string inline storage;
- copy-on-write string builders;
- mutation;
- parser-specific string special cases.

## Semantic invariants

The change must preserve:

- string immutability;
- rune-index semantics;
- Unicode code-point ordering;
- invalid UTF-8 behavior equivalent to Go `[]rune` conversion where such host
  strings reach Aiki;
- index bounds faults;
- `first`, `length`, `ord`, and comparison results;
- Aiki-visible string representation and type.

## Realization

`value.String` remains:

```go
type String struct {
    Val string
}
```

No new physical state is introduced.

Observation methods walk the UTF-8 string directly:

```text
RuneAt(i)
RuneLen()
FirstRune()
CompareRunes(other)
```

These methods may scan bytes but do not allocate a whole `[]rune`.

This deliberately trades temporary allocation for direct immutable observation.

## Performance hypothesis

Current random rune access:

```text
O(n) scan + O(n) allocation
```

Initial corrected realization:

```text
O(n) scan + O(1) allocation
```

where the returned Aiki `Rune` wrapper remains the semantic result.

For repeated indexing the asymptotic scan cost remains. That is intentional.

If post-fix CPU profiling shows repeated rune scanning as the next dominant
cost, a rune-offset index/cache may be proposed separately. It is not admitted
preemptively.

## Systems witness distinction

The PDP-11 allocation survey has a different dominant family:

```text
Parser.recordFailure        604.67 MB   60.05%
```

with parser terminal/error construction dominating allocation objects.

That is recorded as a separate parser-realization concern. It is not folded into
this string proposal merely because it appears in the same survey.

## Delivery plan

This is a one-gate wave.

### Tranche A — implement continuously

1. establish observation authority in `value.String`;
2. migrate the measured evaluator indexing site;
3. migrate same-class observation primitives;
4. add Unicode/equivalence/allocation tests;
5. add focused witness;
6. update durable state.

No user stop is required during this tranche unless a semantic contradiction
appears.

### Critical Gate — survival decision

Run:

```text
make validate
aiki profile --counts extra/profiling/string-observation.ai
aiki profile --counts extra/profiling/selfhost-three-level.ai
AIKI_PDP_PERF_SCALE=10 aiki profile --counts experiments/004-v6-emulator/experiment/diagnostics/cut7_perf_loop.ai
```

The implementation survives if:

- whole-tree validation passes;
- semantic counts remain unchanged;
- self-host allocation bytes fall materially;
- self-host elapsed time is favorable or not materially worse;
- PDP remains materially flat;
- the dominant `evalStringIndex` allocation family disappears when the
  allocation survey is repeated.

If allocation disappears but CPU time materially worsens from repeated scans,
do not hide that result. That is the evidence threshold for considering a
separate adaptive rune-index representation.

## Rejection rule

Do not add a rune cache, alternate string representation, rope, or builder to
rescue this proposal.

This proposal answers one question only:

> Can immutable string observation stop materializing whole strings?

If yes, close it.

If no, record why and open a separately justified representation proposal.

## Expected architectural result

The successful result should be expressible as:

> **The string is immutable. Observation need not materialize it.**


## Focused witness — PASSED

Authoritative focused run:

```text
aiki profile --counts extra/profiling/string-observation.ai
```

Result:

```text
1538

Aiki semantic work
  arithmetic   201538
  comparison   200001
  call         200006
  iteration    100000
  index        100000

Go substrate realization
  elapsed      385.940546ms
  alloc_bytes  106956928
  mallocs      4020176
  gc_cycles    8
```

The expected hit count is 1,538: the source contains one target rune in a
65-rune cycle over 100,000 iterations.

The witness also exposed and documented an important semantic distinction:

```text
"猫"          -> string
first("猫")   -> rune
text[i]       -> rune
```

Accordingly, an indexed string result must be compared with another rune, not
with a one-character string.

The focused witness therefore passes both semantic and realization intent.
The remaining gate is the post-change allocation survey of the three-level
self-host and PDP regression witness.

No alternate string representation is admitted at this stage.


## Static closure audit

After the observation change, whole-string `[]rune` conversions remain only
where the operation actually materializes/transforms rune content or outside the
string-observation path.

Examples intentionally remaining include:

- `rest(string)` — constructs a new suffix string;
- substring — constructs a new substring;
- reverse — transforms rune order;
- `chars` — explicitly materializes a list of runes;
- trim/search helpers — separate operations not selected by the measured
  `evalStringIndex` pathology;
- rune-literal decoding;
- symbol ordering.

The following observation paths are now protected by invariant:

- evaluator string indexing;
- `first(string)`;
- `length(string)`;
- `ord(string)`;
- `String.RuneAt`;
- `String.RuneLen`;
- `String.FirstRune`;
- rune-wise natural string comparison.

This proposal does not prohibit all `[]rune` conversions. It prohibits
whole-string materialization where the semantic operation is only observation.

## Final evidence gate

Status: **PENDING**

The old alpha-36 self-host allocation-space profile was:

```text
evalStringIndex     6171.30 MB   73.20%
NewCallEnv           931.20 MB   11.04%
evalCallArgs         210.01 MB    2.49%
Parser.recordFailure 192.56 MB    2.28%
NewNumberFromString  159.51 MB    1.89%
```

If `evalStringIndex` disappears after this tranche, those residual sites become
the next survey candidates. They are deliberately not changed inside this
proposal.

The string project closes without an alternate representation if:

1. `evalStringIndex` no longer materially allocates;
2. self-host allocated bytes fall approximately in line with removal of the old
   6.17 GB site;
3. self-host elapsed time is favorable or not materially worse;
4. PDP remains materially flat;
5. `make validate` passes.

Only if allocation disappears but repeated rune scanning becomes a dominant CPU
site may a later proposal consider a cached/indexed string representation.


## Gate execution

The complete final gate is intentionally one command:

```sh
extra/profiling/string-observation-gate.sh
```

It runs:

1. `make validate`;
2. the focused mixed-width Unicode observation witness;
3. the post-fix self-host allocation survey;
4. the PDP-11 10x regression survey;
5. packaging of all evidence under `/tmp`.

Generated profiling output is not repository content and must not be admitted to
treecheck or committed.
