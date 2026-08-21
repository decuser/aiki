# Proposal: Adaptive Persistent List Representation

## Status

**COMPLETE**

## Baseline

Baseline: `b829e2c`, branch `runtime-optimization`.

Authoritative three-level self-host baseline:

```text
Aiki semantic work
  arithmetic   1289910
  comparison   1911310
  call         8701738
  iteration    1005458
  index        1443452

Call realization
  user_entry           4664042
  substrate            3892497
  tail_reuse           143766
  tail_env_reuse       143766

Go substrate realization
  elapsed      14.427563993s
  alloc_bytes  25005019784
  mallocs      50560133
  gc_cycles    829
```

Authoritative PDP-11 Cut 7 10x regression witness:

```text
Go substrate realization
  elapsed      1.562006621s
  alloc_bytes  1040319496
  mallocs      16463419
  gc_cycles    33
```

The call/allocation tranche that produced this baseline is complete. This
proposal begins a new optimization boundary.

## Summary

Aiki has one semantic list type: `list`.

A list is persistent. Operations such as `append` return a new list value and do
not alter the semantic value of the input list. That semantic rule does not
require every list operation to allocate and copy a complete new backing slice.

The current `append` implementation always performs:

```go
newElems := make([]value.Value, len(list.Elements)+1)
copy(newElems, list.Elements)
newElems[len(list.Elements)] = item
```

Repeated construction such as:

```aiki
let xs = []
xs = append(xs, a)
xs = append(xs, b)
xs = append(xs, c)
...
```

therefore realizes backing arrays of lengths 1, 2, 3, ... N. The semantic
result is linear growth; the physical realization is quadratic copied-slot
volume.

This proposal separates **persistent list semantics** from **list
representation**.

> **The list is immutable. The representation is negotiable.**

The initial adaptive representation has two hidden states:

```text
                        Aiki list
                            |
                persistent semantic value
                            |
                  +---------+---------+
                  |                   |
               flat list       frontier-backed list
                  |                   |
                  | append            | append at frontier
                  +---- promotion --->| amortized growth
                                      |
                                      | append from
                                      | historical prefix
                                      v
                                  fork backing
```

These are implementation states, not user-visible list kinds.

The initial proposal intentionally does **not** add ropes, trees, inline-small
lists, mutable builders, or a new programmer-visible collection type. Those
would require separate evidence.

## Architectural principle

For lists:

```text
Aiki meaning
    list

semantic contract
    ordered persistent sequence, optionally shaped

runtime realization
    flat immutable storage or shared frontier backing
```

Representation may affect cost but must not affect:

- `type`;
- element order;
- length;
- indexing;
- equality;
- truth;
- inspection/printing;
- shape;
- iteration;
- alias behavior;
- concurrency behavior.

## Semantic invariants

### 1. Persistence

If:

```aiki
let a = [1, 2]
let b = append(a, 3)
```

then `a` remains `[1, 2]` forever and `b` is `[1, 2, 3]`.

No optimization may make an older list observe elements appended to a newer
list.

### 2. Branching persistence

If:

```aiki
let a = [1, 2]
let b = append(a, 3)
let c = append(a, 4)
```

then:

```text
a == [1, 2]
b == [1, 2, 3]
c == [1, 2, 4]
```

The second append is a branch from a historical prefix and must fork rather
than overwrite the path that produced `b`.

### 3. One list type

All representations remain `list`. Representation is not inspectable from Aiki.

### 4. Representation-independent observation

Length, element access, iteration, equality, shaped-field access, display,
hashing where applicable, and all library operations observe the same logical
sequence regardless of representation.

### 5. Shape preservation

List representation is independent of shape metadata. Existing shaped-list
behavior is unchanged.

### 6. No mutation authority leaks

Runtime consumers may not mutate a logical list through a raw backing slice.
Any compatibility view retained during migration is read-only by invariant.

### 7. Spawn/channel safety

Aiki values can cross concurrent execution boundaries. Two concurrent appends
from aliases of the same frontier must remain race-free and semantically
persistent.

Frontier mutation therefore requires internal synchronization or an equivalent
publication protocol. This proposal does not rely on single-threaded use,
reference counts, or compiler escape analysis.

### 8. No semantic dependence on capacity

Backing capacity, frontier position, fork count, and promotion state are
realization facts only.

## Representation model

### Flat list

A flat list is the cheap representation for values that are created and merely
read.

Conceptually:

```text
flat
  logical elements: [a b c]
```

It requires no append-specific synchronization metadata.

A flat list may share an immutable subslice under existing operations such as
`rest`; therefore its spare capacity, if any, is never treated as mutation
authority.

### Frontier backing

On append from a flat list, the result may promote to a frontier backing:

```text
backing:
  slots: [a b c d _ _ _ _]
  used:               4

old flat:
  [a b c]

new list:
  backing B
  logical length 4
```

Promotion copies the logical flat prefix into runtime-owned backing storage.
This one-time copy establishes mutation authority over unused capacity without
making assumptions about aliases to the original flat slice.

A frontier backing owns:

- runtime-owned element storage;
- a published frontier (`used`);
- growth capacity;
- synchronization for competing appends.

A list wrapper owns its logical length and shape.

### Frontier append

For a frontier-backed list `L`:

```text
if L.length == backing.used
```

then `L` is at the current frontier.

Appending may reserve/write the next unused slot and return a new wrapper with
logical length `L.length + 1`.

Older wrappers still have shorter logical lengths and cannot observe the new
slot.

If capacity is exhausted, the backing may grow geometrically. Existing list
wrappers must continue to observe their exact logical prefixes.

### Historical fork

If:

```text
L.length < backing.used
```

then `L` is a historical prefix.

Appending to it must not write the already-used next slot. Instead:

1. allocate a new frontier backing;
2. copy exactly `L`'s logical prefix;
3. append the new element;
4. return a wrapper over the new frontier.

This is the promotion/fork escape path that preserves persistent branching.

### Concurrency

The simple correct initial realization is a lock inside frontier backing.

Readers of published logical prefixes remain read-only. Append/fork operations
serialize frontier state transitions. Two concurrent appends to the same
frontier therefore produce one linear successor and one branch/fork (or an
equivalent persistent outcome), never a data race or overwrite.

A more elaborate lock-free publication scheme is explicitly out of scope unless
profiling later proves the lock itself material.

## Why this is analogous to Adaptive Number

Adaptive Number established:

> **The number is exact. The representation is negotiable.**

Its representations survive only when they preserve exact semantics and earn
their place under real workloads.

Adaptive List applies the same rule:

> **The list is immutable. The representation is negotiable.**

For Number, overflow or uncertified arithmetic promotes to a representation
that can preserve exactness.

For List, an append that cannot safely extend the current frontier forks to a
representation/backing that preserves persistence.

Neither system exposes promotion to the programmer.

## Semantic authority

The baseline audit found 116 direct `List.Elements` uses across production/test
surfaces but only one production mutation through an existing list. Tranche A
does not require a wholesale read migration. `Elements` remains a compatibility
view of the immutable logical prefix while adaptive ownership metadata is
private. Representation authority first means that mutation and frontier
extension are centralized; read sites may migrate to `Len`, `At`, iteration, or
`LogicalElements` when doing so removes representation pressure rather than as
ceremony.

The authority surface should remain small and unsurprising, conceptually:

```text
Len()
At(i)
ForEach / logical iteration
LogicalElements()       // read-only/materialized boundary where truly required
Append(value)
Prepend(value)
Slice/Rest
Shape
```

Exact names are implementation decisions.

The migration rule is:

> Runtime code may ask a List for its logical sequence; it may not acquire
> mutation authority over the physical representation.

The proposal does not require an artificial abstraction ceremony where direct
iteration is harmless. The purpose of authority is to make alternate
representations possible and to prevent accidental backing mutation.

## Initial scope

Included:

- list representation authority;
- flat representation;
- frontier-backed representation;
- flat-to-frontier promotion;
- amortized frontier append;
- historical-prefix fork;
- synchronized concurrent append;
- shape preservation;
- representation profiling;
- semantic/property/differential tests;
- self-host evidence;
- PDP-11 regression evidence;
- invariants and documentation.

Not included:

- mutable Aiki lists;
- user-visible builders;
- vectors as a second collection type;
- ropes;
- persistent vector trees;
- small-list inline representation;
- optimized prepend;
- optimized concat;
- optimized `rest`/slice views beyond what falls out naturally;
- hash redesign;
- parser-specific special cases.

Any of those require separate evidence after this proposal.

## Instrumentation

Profiling should report realization facts, not semantic kinds.

The initial counters are append-focused so they measure the mechanism under
test without pretending to account for every flat-list constructor in the tree:

```text
List realization
  frontier_promoted
  frontier_extended
  frontier_grown
  frontier_forked
  elements_copied
  backing_slots_allocated
```

Where practical, record copied element slots:

```text
  elements_copied
```

This is the critical metric. The hypothesis is not merely that `append` causes
many allocations; it is that repeated persistent append currently causes
excessive copied-slot volume.

These counters are profiling-only and must impose negligible cost when profiling
is disabled.

## Correctness tests

The implementation must include targeted tests for:

- append preserves input;
- long linear append chain;
- branch from every historical depth;
- branch-after-grow;
- repeated branching;
- append to a `rest`-derived list (promotion must copy the logical prefix);
- append to the original list after deriving `rest`;
- aliases through `rest`;
- shaped lists;
- nested lists;
- equality across flat/frontier representations;
- inspection across representations;
- index and length boundaries;
- empty list;
- prepend followed by append;
- append followed by rest/slice;
- concurrent appends from the same frontier;
- concurrent appends from historical aliases;
- spawn/channel transfer of lists;
- randomized differential comparison against a simple immutable reference
  implementation.

The randomized differential test is load-bearing.

## Performance hypothesis

The expected winning workload is incremental construction:

```aiki
let xs = []
while ... {
    xs = append(xs, value)
}
```

Current copied-slot volume is approximately:

```text
1 + 2 + ... + N = O(N^2)
```

A frontier-backed linear chain should approach:

```text
O(N)
```

allocated/copied slots with geometric growth.

Branching remains proportional to the copied historical prefix because
persistence genuinely requires a fork there.

The optimization is accepted only if real workload evidence shows material
benefit.

## Workloads

### Primary pressure witness

`extra/profiling/selfhost-three-level.ai`

This is the architectural pressure workload because the self-host parser and
interpreter construct many lists incrementally.

Authoritative baseline:

```text
elapsed      14.427563993s
alloc_bytes  25005019784
mallocs      50560133
gc_cycles    829
```

### Systems regression witness

`experiments/004-v6-emulator/experiment/diagnostics/cut7_perf_loop.ai`

Run at:

```text
AIKI_PDP_PERF_SCALE=10
```

Authoritative baseline:

```text
elapsed      1.562006621s
alloc_bytes  1040319496
mallocs      16463419
gc_cycles    33
```

The PDP workload is a regression witness, not the architectural driver.

### Focused list witness

Add one durable profiling workload that constructs:

1. a long linear append chain;
2. branches from historical prefixes;
3. shaped and unshaped lists.

It should make promotion/extension/fork counts obvious without relying on
wall-clock timing.

## Delivery strategy: fewest responsible stops

This proposal is intentionally delivered in **two implementation tranches**,
not a succession of user-gated micro-cuts.

Internal cuts may exist for source-control clarity, but they are not user stops.

### Tranche A — semantic authority + adaptive representation

Do continuously:

1. audit all `List.Elements` reads/writes and construction sites;
2. establish List semantic authority;
3. add flat and frontier representations;
4. implement promotion, geometric growth, and historical fork;
5. make frontier transition concurrency-safe;
6. migrate runtime consumers;
7. add randomized differential, aliasing, shape, spawn/concurrency, and
   representation tests;
8. add profiling counters;
9. update invariants and focused profiling witness.

#### Critical Gate 1

Stop only after Tranche A is complete.

Required gate:

```text
focused value/evaluator/substrate tests
concurrency/race-relevant tests where available
make validate
```

Gate 1 answers only:

> Does the adaptive representation preserve Aiki list semantics and whole-tree
> invariants?

No performance judgment is made before this gate.

If Gate 1 fails, fix continuously until either it passes or a genuine semantic
design contradiction requires user input.

### Tranche B — evidence + reconciliation

After Gate 1 passes, continue without intermediate stops:

1. run the focused list witness;
2. run the three-level self-host witness;
3. run PDP-11 Cut 7 10x;
4. inspect realization counters;
5. remove dead/unused representation machinery;
6. reconcile docs/invariants;
7. run final whole-tree validation.

No small-list representation or second optimization family is introduced in
this tranche. The purpose is to decide whether the frontier design itself earns
its place.

#### Critical Gate 2 — final evidence gate

Stop once with the complete evidence.

The representation survives only if:

- semantic counts remain unchanged;
- differential/property/concurrency tests remain green;
- self-host copied-slot/allocation volume falls materially;
- self-host elapsed time is favorable or at minimum not materially worse;
- PDP has no material regression;
- `make validate` is green.

If the adaptive path does not earn its place, revert it cleanly rather than
adding compensating complexity.

## User-input policy during implementation

The implementation should proceed without asking for routine approvals,
microbenchmark runs, or per-cut confirmation.

User input is required only when one of these occurs:

1. a semantic ambiguity not resolved by this proposal;
2. a correctness/invariant failure that presents two materially different
   language-design choices;
3. a baseline mismatch that makes continued implementation unsafe;
4. Critical Gate 1;
5. Critical Gate 2.

Compilation errors, ordinary test failures, instrumentation mistakes, and
implementation bugs are work to fix, not reasons to stop and ask.

## Rejection criteria

Reject or simplify the design if:

- frontier synchronization costs erase the allocation-volume benefit;
- historical-prefix forks materially dominate frontier extension, or copied-element volume remains high enough to erase the expected linear-construction benefit;
- representation authority materially complicates ordinary list operations
  without measured benefit;
- self-host allocation bytes do not improve materially;
- PDP/system workloads regress materially;
- concurrency semantics require assumptions the language does not guarantee.

Do not rescue a weak representation by adding ropes, builders, trees, or
compiler escape analysis inside this proposal. If branching from historical
prefixes is common in the self-host workload, the realization counters must
show it directly; reject or simplify rather than compensating with another
persistence structure.

## Success criterion

The proposal succeeds if ordinary persistent Aiki list construction can use an
amortized append path while branching, aliases, shapes, and concurrency remain
semantically indistinguishable from the current immutable-copy implementation.

The architectural result should be expressible in one sentence:

> **The list is immutable. The representation is negotiable.**


## Tranche A implementation state — 2026-08-20

Implemented continuously from baseline `b829e2c`:

- private synchronized frontier backing with logical-prefix wrappers;
- flat-to-frontier promotion by prefix copy;
- amortized frontier extension and geometric growth;
- historical-prefix fork;
- explicit `rest`-derived flat alias behavior;
- no direct production mutation of `List.Elements` outside value authority;
- lightweight probe-aware `_append` path that does not require full `EvalContext`;
- append-focused realization counters;
- randomized differential, historical-branch, shape, rest/original alias, and
  concurrent-append tests;
- invariant rejecting direct `List.Elements` assignment outside value authority;
- focused `extra/profiling/adaptive-list.ai` witness.

Local environment limitation: the container cannot fetch the required Go 1.24
toolchain or external dependencies. A disposable Go 1.23 compatibility run with
local `readline`/`flock` stubs passed value, evaluator, invariant, focused
substrate, and value race tests. The only full substrate failure under the stub
was the known file-lock exclusivity test, which the fake flock implementation
cannot validate. This is directional evidence only; Critical Gate 1 remains the
authoritative Go 1.24 `make validate` gate.


## Critical Gate 1 — PASSED

Authoritative Go 1.24 validation on the project tree passed:

```text
make validate                         PASS
go test -race adaptive-list tests     PASS
```

Focused realization witness:

```text
aiki profile --counts extra/profiling/adaptive-list.ai
```

Observed:

```text
frontier_promoted        3
frontier_extended        100000
frontier_grown           15
frontier_forked          1
elements_copied          131070
backing_slots_allocated  262152

elapsed      240.729093ms
alloc_bytes  77831816
mallocs      2300308
gc_cycles    5
```

The focused witness demonstrated the intended economics:

- 100,000 linear frontier extensions;
- only one historical-prefix fork;
- 15 geometric grows;
- persistence examples remained correct;
- copied-element volume stayed linear-scale rather than quadratic.

Gate 1 therefore established correctness and representation behavior on the
authoritative tree.

## Critical Gate 2 — PASSED

### Three-level self-host

Pre-list baseline at `b829e2c`:

```text
elapsed      14.427563993s
alloc_bytes  25005019784
mallocs      50560133
gc_cycles    829
```

Adaptive persistent-list result:

```text
List realization
  frontier_promoted        52928
  frontier_extended        209407
  frontier_grown           4039
  frontier_forked          0
  elements_copied          274328
  backing_slots_allocated  760368

Go substrate realization
  elapsed      10.619609147s
  alloc_bytes  8901681936
  mallocs      50401888
  gc_cycles    267
```

Relative to the pre-list baseline:

- elapsed improved about 26.4%;
- allocated bytes fell about 64.4%;
- malloc count was essentially flat, confirming that the dominant problem was
  allocation volume rather than allocation frequency;
- GC cycles fell about 67.8%;
- `frontier_forked` was zero in the three-level self-host workload.

The workload therefore strongly supports the frontier hypothesis: self-host
list construction is overwhelmingly linear-growth rather than historical
branching.

### PDP-11 regression witness

Pre-list baseline:

```text
elapsed      1.562006621s
alloc_bytes  1040319496
mallocs      16463419
gc_cycles    33
```

Adaptive persistent-list result:

```text
elapsed      1.569385077s
alloc_bytes  1040367792
mallocs      16463466
gc_cycles    34
```

The difference is materially flat for the systems regression witness. Semantic
and call-realization counts remained unchanged.

### Final whole-tree gate

```text
make validate    PASS
```

Critical Gate 2 therefore passed.

## Completion decision

Completed 2026-08-21.

The adaptive persistent-list representation **earns its place**.

The surviving design is intentionally limited to:

- flat persistent lists;
- private synchronized frontier backing;
- flat-to-frontier promotion by copying the logical prefix;
- amortized frontier extension and geometric growth;
- historical-prefix fork;
- immutable logical-prefix observation through existing list semantics;
- profiling counters for promotion/extension/growth/fork/copy volume.

The proposal explicitly does **not** admit:

- ropes;
- persistent trees;
- user-visible builders;
- small-list inline representations;
- compiler escape-analysis machinery;
- parser-specific list special cases.

No compensating structure is needed. The measured workload that motivated the
proposal has zero historical forks and receives a 64.4% reduction in allocation
bytes.


The programmer-facing semantic driver is immutability; persistence is the
branch-preservation property that append must maintain.

The final architectural statement is:

> **The list is immutable. The representation is negotiable.**
