# Adaptive Number Cut 0 Audit

Status: **GATED**

Controlling proposal: `proposals/completed/adaptive-exact-number-representation.md`

Baseline commit: `cbddead`

## Purpose

Record the current representation coupling before changing `Number`. Cut 0 is an
inventory and baseline-measurement cut only; it does not alter runtime behavior.

## Current representation authority

`engine/semantics/value/value.go` currently defines:

```go
type Number struct {
    Val *big.Rat
}
```

That exported field makes `big.Rat` not merely the implementation of Number but a
widely consumed representation API. The adaptive-number project must first make
Number the semantic authority before introducing alternate representations.

## Production representation coupling

The audit found direct `Number.Val` / `big.Rat`-representation dependence in 26
production Go files.

### Semantic core

- `engine/semantics/value/value.go` — storage, construction, inspection, deep equality.
- `engine/semantics/value/order.go` — number ordering reaches through to `big.Rat.Cmp`.
- `engine/semantics/evaluator/operators.go` — `+`, `-`, `*`, `/` construct `big.Rat` directly.
- `engine/semantics/evaluator/expressions.go` — unary negation constructs `big.Rat` directly.
- `engine/semantics/evaluator/patterns.go` — number equality reaches through to `big.Rat.Cmp`.
- `engine/semantics/evaluator/access.go` — indexing reaches through `IsInt`/`Num`.

### HAL/substrate consumers

Direct representation access appears in:

- `builtins_bits.go`
- `builtins_bytes.go`
- `builtins_canvas.go`
- `builtins_canvas_accessors.go`
- `builtins_convert.go`
- `builtins_file.go`
- `builtins_hash_provider.go`
- `builtins_io.go`
- `builtins_list.go`
- `builtins_machine_ffi.go`
- `builtins_math.go`
- `builtins_network.go`
- `builtins_store.go`
- `builtins_store_ffi.go`
- `builtins_string_provider.go`
- `builtins_system.go`
- `builtins_time.go`
- `builtins_trig.go`
- `builtins_type.go`
- `byte_normalize.go`

The uses fall into a small semantic set that should become Number-owned APIs:

```text
exact arithmetic
comparison/equality
sign/zero
integer predicate
signed/unsigned integer conversion
numerator/denominator access when mathematically required
host binary64 conversion at explicit boundaries
construction from large integers / exact rational components
inspection/hash canonicalization
```

This is favorable: the representation leakage is broad in files but narrow in
kind.

## Test coupling

Ten test files assert or consume the concrete representation directly:

- `engine/runtime/hal/substrate/builtins_bits_test.go`
- `engine/runtime/hal/substrate/builtins_canvas_accessors_test.go`
- `engine/runtime/hal/substrate/builtins_machine_ffi_test.go`
- `engine/runtime/hal/substrate/builtins_network_test.go`
- `engine/runtime/hal/substrate/builtins_store_test.go`
- `engine/runtime/hal/substrate/m5_affordances_test.go`
- `engine/runtime/hal/substrate/substrate_test.go`
- `test/boundary/hal_prelude_user_gating_test.go`
- `test/contract/exact_number_behavior_test.go`
- `test/property/value_properties_test.go`

The strongest representation assertions are currently wrong for the accepted
architecture:

- `TestNumbersAreBigRat` treats representation as contract.
- property documentation says number arithmetic always produces `*big.Rat`.
- tests construct `Number{Val: ...}` directly.

Cut 1 should replace these with semantic-value assertions while retaining the
existing exact behavior.

## Architectural invariant conflict identified

`test/invariant/exact_number_architecture_test.go` currently bans all float types
and `Float64`/`SetFloat64` conversion paths under `engine/semantics` and
`engine/syntax`.

That invariant correctly protected Aiki from reintroducing approximate ordinary
arithmetic, but it states the protection as a representation prohibition. It will
conflict with the accepted exact-binary64 carrier in Cut 4.

Disposition:

- do **not** weaken it during Cut 1;
- keep float conversion outside the semantic core while Number authority is first
  established;
- replace it in Cut 4/Cut 8 with a semantic invariant that permits an exact
  binary64 carrier but forbids rounded ordinary arithmetic.

## Host math boundary

The existing design decision D1 is compatible with the new proposal. Host-backed
math computes in binary64, then the returned finite bit pattern is converted to
an Aiki exact rational value. The new proposal changes only the hidden storage
choice: Cut 4 may retain those bits directly as an exact dyadic carrier rather
than immediately expanding them to `big.Rat`.

## Machine-domain boundary

`lib/machine/ffi` and its substrate implementation are not candidates for
conversion back into ordinary Number arithmetic. The fixed-width machine domain
owns wraparound machine words/bytes/addresses. Adaptive Number and machine values
remain separate semantic boundaries.

## Documentation coupling

Representation-specific claims also occur in:

- `docs/adding-to-aiki.md` — example code reaches through `Number.Val`.
- `test/property/README.md` — claims `*big.Rat` as the Number property.
- `test/invariant/README.md` — describes the no-float representation invariant.
- historical session/proposal material — retained as history unless a current
  statement is misleading.

`docs/decisions.md` D1 already states the important semantic boundary correctly
and should be refined, not reversed, during documentation reconciliation.

## Baseline benchmark surface

`engine/semantics/value/number_baseline_benchmark_test.go` records the current
`big.Rat` realization for:

```text
integer construction
integer addition
rational addition
integer multiplication
integer comparison
rational division
```

Run authoritatively with:

```bash
go test ./engine/semantics/value -run '^$' \
  -bench '^BenchmarkNumberBaseline' -benchmem -count=3
```

The authoritative Go 1.24 baseline supplied from the project host measured:

```text
construct integer       ~78 ns/op     40 B/op   5 allocs/op
add integer            ~190 ns/op    216 B/op   7 allocs/op
add rational           ~193 ns/op    216 B/op   7 allocs/op
multiply integer       ~128 ns/op    120 B/op   5 allocs/op
compare integer         ~67 ns/op      96 B/op   2 allocs/op
divide rational        ~136 ns/op    120 B/op   5 allocs/op
```

These figures are diagnostic only. The authoritative Go 1.24 run must be retained
before Cut 0 is `GATED`.

## Cut 0 gate state

**GATED.** Authoritative Go 1.24 evidence supplied on 2026-08-20:

```text
go test ./engine/semantics/value ./engine/semantics/evaluator ./test/contract ./test/invariant
    PASS

go test ./engine/semantics/value -run '^$' -bench '^BenchmarkNumberBaseline' -benchmem -count=3
    PASS; baseline retained above
```

Cut 1 therefore proceeds from a validated semantic and performance baseline.
