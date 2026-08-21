# Adaptive Exact Number — Evidence Record

Controlling proposal: `proposals/completed/adaptive-exact-number-representation.md`

Date: 2026-08-20

This record captures the acceptance evidence for Aiki's adaptive hidden Number
representations. The semantic boundary remains one exact Aiki `number`; the
representations below are implementation realizations only.

## Baseline

Authoritative host: Linux/amd64, AMD Ryzen 7 7840HS, Go 1.24 project toolchain.

Before adaptive representation, ordinary Number operations universally paid the
`big.Rat` realization cost:

```text
integer construction   ~78 ns    40 B    5 allocs
integer addition      ~190 ns   216 B    7 allocs
rational addition     ~193 ns   216 B    7 allocs
integer multiply      ~128 ns   120 B    5 allocs
integer comparison     ~67 ns    96 B    2 allocs
rational division     ~136 ns   120 B    5 allocs
```

## Adaptive microbenchmarks

```text
integer construction   ~22 ns    32 B    1 alloc
integer addition        ~28 ns    32 B    1 alloc
rational addition       ~41 ns    32 B    1 alloc
integer multiply        ~29 ns    32 B    1 alloc
integer comparison     ~4.7 ns     0 B    0 alloc
rational division       ~38 ns    32 B    1 alloc
binary carrier          ~21 ns    32 B    1 alloc
```

Binary-path evidence:

```text
certified binary add       ~22 ns    32 B     1 alloc
certified binary multiply  ~27 ns    32 B     1 alloc
common exact fallback      ~52 ns    32 B     1 alloc
true big fallback         ~1.87 us  2209 B    26 allocs
```

A failed binary certificate first re-expresses the exact dyadic value as a
compact rational when possible. Arbitrary precision is therefore the escape
path rather than the default fallback.

## Three-level self-host witness

Workload:

```text
host Aiki -> selfhost -> selfhost -> 1 + 2 * 3
```

Result: `9`, preserving Aiki's left-to-right evaluation.

```text
Aiki semantic work
  arithmetic   1289910
  comparison   1911310
  call         8701738
  iteration    1005458
  index        1443452

Number arithmetic realization
  small_integer        1122252
  compact_rational     0
  binary_carrier       0
  big_rational         0

Go substrate realization
  elapsed      19.394718734s
  alloc_bytes  28291183320
  mallocs      110603982
  gc_cycles    899
```

This is strong coverage for the small-integer representation because the
workload is a complete lexer/parser/evaluator/runtime stack recursively
interpreting itself, not an arithmetic microbenchmark.

## Experiment 004 PDP-11 witness

Cut 7, 10x, 7,680 guest instructions:

```text
Aiki semantic work
  arithmetic   26008
  comparison   41339
  call         541586
  iteration    7906
  index        258
  store_read   9
  store_write  119

Number arithmetic realization
  small_integer        20834
  compact_rational     0
  binary_carrier       0
  big_rational         0

Go substrate realization
  elapsed      1.923894192s
  alloc_bytes  1301999120
  mallocs      21122133
  gc_cycles    38
```

The immediately preceding same-work profile, before adaptive Number, was about
2.534 s, 1.667 GB allocated, and 26.47 M mallocs with the same 541,586 Aiki
calls. Adaptive Number therefore improved realization cost without changing the
PDP execution algorithm. The PDP word/address interior remains in the separate
fixed-width machine FFI domain; the two boundaries are complementary.

## Rational witness

```text
outputs
  0
  3/10
  1

Aiki semantic work
  arithmetic   500005
  comparison   100001
  iteration    100000

Number arithmetic realization
  small_integer        300001
  compact_rational     200004
  binary_carrier       0
  big_rational         0

elapsed      223.62189ms
alloc_bytes  49623832
mallocs      2100274
```

Compact rational is therefore not speculative machinery: it carries substantial
exact-rational work without arbitrary precision.

## Host-math witness

The witness imports `math/ffi`, so each finite host binary64 return crosses into
Aiki as the exact dyadic rational denoted by those bits. Call returns and
arithmetic results are measured separately.

```text
Number arithmetic realization
  small_integer        200002
  compact_rational     54486
  binary_carrier       145436
  big_rational         76
  binary_certified     245435
  binary_fallback      54562
  promoted_big         76

Number call-return realization
  small_integer        4
  compact_rational     0
  binary_carrier       399998
  big_rational         0

elapsed      643.673007ms
alloc_bytes  295007288
mallocs      7608645
```

Of 299,997 binary arithmetic attempts, 245,435 (~81.8%) were certified exact.
Only 76 (~0.025%) promoted to arbitrary precision. Most failed certifications
fell exactly into compact rational.

## Representation decision

All three adaptive fast representations survive their evidence gate:

- small integer: dominant in ordinary interpreter and systems bookkeeping;
- compact rational: substantial exact-rational coverage and the normal binary
  fallback realization;
- binary carrier: dominant at host-math return boundaries and retained through
  most measured carrier arithmetic under exact certification.

`big.Rat` remains the necessary arbitrary-precision escape representation.

Certified binary division is **not admitted**. The existing exact fallback is
cheap enough and the measured workloads do not justify adding another proof
path.

## Resulting invariant

> The number is exact. The representation is negotiable.

The binary64 carrier does not add a float type to Aiki and does not authorize
rounded ordinary arithmetic. The fixed-width machine FFI domain remains a
separate systems-programming boundary and does not re-enter ordinary Number in
machine interiors.


## Whole-tree completion gate

`make validate` passed on the authoritative Go 1.24 project host after the
profiling subtree was reconciled with treecheck.

Adaptive Number is therefore complete. The representation set is frozen at the
surviving evidence-gated realizations documented above. Further numeric
complexity requires a new proposal and new workload evidence rather than
continuing this effort.
