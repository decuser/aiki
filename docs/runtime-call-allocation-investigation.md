# Runtime Call and Allocation Investigation

Status: **ACTIVE**

## Purpose

Adaptive Number removed a material numeric cost without changing Aiki's semantic
number boundary. The remaining whole-program profiles now point much more
strongly at call and allocation realization.

This investigation measures that surface before changing function semantics.

## Baseline witnesses

### Three-level selfhost

`extra/profiling/selfhost-three-level.ai`:

```text
call         8,701,738
elapsed      19.394718734s
alloc_bytes  28,291,183,320
mallocs      110,603,982
gc_cycles    899
```

The same workload records 1,122,252 small-integer arithmetic results and no
rational/binary/big results. Number representation is therefore not a plausible
explanation for most remaining allocation.

### PDP-11 Cut 7 stress witness

10x / 7,680 guest instructions:

```text
call         541,586
elapsed      1.923894192s
alloc_bytes  1,301,999,120
mallocs      21,122,133
gc_cycles    38
```

The emulator deliberately keeps PDP-width arithmetic in `machine/ffi`; ordinary
Number realization is again almost entirely small integer.

## Initial call-realization audit

An interpreted user-function entry currently constructs a new `Env`. Before this
cut, every call environment eagerly created both:

- a value-binding map; and
- a shape-definition map.

The shape map is unnecessary for the overwhelming majority of calls. Child
environments also paid the same eager-map policy even when they introduced no
bindings.

Arguments are additionally materialized as slices at call sites. Parameter
bindings are then copied into the call environment. Rest parameters allocate a
List wrapper when present.

These costs are realization choices, not Aiki function semantics.

## Investigation Cut 1

This cut makes child/call environment maps lazy and extends profiler output with
call realization:

```text
Call realization
  user_entry
  substrate
  tail_reuse
```

`user_entry` counts actual interpreted function-frame entries. `substrate`
counts callable crossings into the HAL/substrate. `tail_reuse` counts semantic
calls realized by reusing the current interpreted frame.

No closure lookup, dynamic stack ownership, authority, argument, or tail-call
semantics change.

## Gate

Run both baseline witnesses with `aiki profile --counts` after rebuilding.

The cut survives only if:

1. focused/full validation stays green;
2. semantic work counts remain unchanged for fixed workloads;
3. allocation or elapsed time improves measurably; and
4. call-realization counts give a useful decomposition of the semantic call
   total.

## Cut 2 — Borrowed call parameters

Status: ACTIVE

The first measured cut removed eager call-scope storage and showed that call-frame
allocation is material:

- three-level selfhost mallocs fell from 110.604M to 105.795M and elapsed from
  19.395s to 18.653s;
- PDP Cut 7 mallocs fell from 21.122M to 20.792M and elapsed from 1.924s to
  1.878s.

Cut 2 removes the ordinary fixed-parameter map entirely. A call environment
borrows the immutable parameter-name slice from the function object and the
argument-value slice from the invocation. Parameter update mutates that internal
argument slot; delete is represented by an inline 64-bit mask so lexical
fall-through remains exact without auxiliary allocation. Calls with more than 64
fixed parameters use the existing map representation.

Rest parameters and actual local bindings retain ordinary map storage. This cut
changes only call-frame realization, not lexical lookup, closure ownership,
authority, stack behavior, or tail-call semantics.

Gate: value-layer binding/update/delete semantics green; authoritative workload
measurement pending.

## Cut 4 — Lazy substrate EvalContext realization

The substrate BuiltinFunc ABI historically accepts `*hal.EvalContext` for every
primitive, even though most native primitives cannot observe evaluator context.
Because `Builtin` implements `ContextCallable`, the evaluator consequently built
an EvalContext, measurement closure, and async-fault plumbing for every substrate
call.

Cut 4 makes context dependence explicit substrate metadata. Registration is
conservative by default: a builtin receives EvalContext unless its registration
site explicitly marks it context-free. The audited language/value native
primitives and Native/FFI provider group use the context-free path, except
`_stack_limit`, `_store_get`, and `_store_set`, which observe evaluator state or
profiling. Host and runtime/service registrations remain conservative in this cut.
Unknown external ContextCallable values also continue to receive context.

This changes realization only; Aiki-visible primitive names and semantics are
unchanged.


## Cut 5 result — GATED

Authoritative PDP-11 Cut 7 10x after Cut 5:

- elapsed: 1.680172198 s
- alloc_bytes: 1,163,524,488
- mallocs: 18,543,556
- gc_cycles: 36

Relative to Cut 4, exact-arity realization removed about 8.9 MB and 54,199
mallocs and improved elapsed time by about 1.3%, with unchanged semantic and
call-realization counts. Cut 5 survives.

## Cut 6 — Allocation-free tail AST traversal

Status: ACTIVE

The interpreted-call path enters `evalTail` for every user function body.
Several tail-position helpers were materializing temporary Go slices from the
already-immutable syntax tree solely to inspect or traverse it:

- wrapper unwrapping built `nonterm` and `terms`;
- block evaluation built `stmts`;
- match evaluation built `cases`;
- pipe evaluation built `parts`.

Cut 6 removes those traversal-only slices. Wrapper inspection tracks one child
in place; blocks keep one pending statement; match pairs are scanned directly;
pipes count and traverse AST children without collecting them.

No Aiki value, scope, evaluation order, tail-call rule, or syntax semantics is
changed. This cut reduces realization scratch around interpreted calls rather
than changing call representation.


## Cut 6 result — PROVISIONAL

Authoritative PDP-11 Cut 7 10x after Cut 6:
- elapsed: 1.707881865 s
- alloc_bytes: 1,157,624,080
- mallocs: 18,481,728
- gc_cycles: 35

Cut 6 removed about 5.9 MB and 61,828 mallocs relative to Cut 5, while the
single short elapsed run regressed about 1.6%. It remains provisional.

## Cut 7 — Bypass disabled profile-label regions

Status: ACTIVE

Normal `aiki profile --counts` runs have pprof labels disabled, but user and
substrate call paths still constructed label-region closures. Cut 7 adds an
optional runtime state query and directly executes the call path when labels are
off. Explicit CPU/profile-label runs retain the existing label behavior.


## Cut 7 result — GATED

Authoritative PDP-11 Cut 7 10x after Cut 7:

- elapsed: 1.631288532 s
- alloc_bytes: 1,084,720,272
- mallocs: 17,916,905
- gc_cycles: 34

Relative to Cut 6, elapsed improved about 4.5%, allocation bytes about 6.3%,
and malloc count about 3.1%, with unchanged semantic and call-realization
counts. Cut 7 survives.

## Cut 8 — Make disabled profile labels fully cold

Status: ACTIVE

Cut 7 bypassed label-region closures when pprof labels are disabled, but the
substrate call branch still eagerly recovered parent semantic labels and built
substrate label metadata before testing whether labels were active.

Cut 8 moves all parent-label recovery, line conversion, and primitive label
construction inside the enabled branch. The normal disabled-label path passes
an empty internal label set to context construction and performs none of that
metadata work.

Explicit CPU/profile-label runs retain the prior behavior.


## Cut 9 — Lazy semantic call detail

Status: ACTIVE

Scalar semantic counting does not retain call-site detail, but `applyFunction`
previously called `Inspect()` on every callee before the probe decided whether
attribution sites were enabled. Named functions and substrate builtins can format
new strings in `Inspect()`, so `--counts` paid that cost millions of times and
discarded the result.

Cut 9 moves call-detail construction behind `AttributionProbe.WantsSites()`.
Attributed profiles retain the existing detail strings; scalar counting records
only the call count and performs no callee inspection.


## Cut 10 — Remove remaining evaluator traversal scratch

Status: ACTIVE

Two ordinary evaluator paths still allocated temporary slices solely to traverse
the immutable syntax tree: unary prefix collection and non-tail match case
collection. Cut 10 walks the existing child array directly. Unary prefixes are
applied by reverse traversal after operand evaluation; match keeps one pending
pattern while scanning pattern/block pairs.

No Aiki value representation, evaluation order, pattern behavior, or call ABI
changes.


## Cut 11 — Pattern matching scratch allocation

Status: ACTIVE

Pattern matching still had two realization-only allocation sources. List
patterns collected nonterminal children into a temporary slice before matching,
and every match case allocated a bindings map even when the pattern contained
no binding names.

Cut 11 traverses list-pattern children directly and allocates a bindings map
only when a recursive allocation-free scan finds a bindable `NAME`. Literal,
wildcard, boolean, and other non-binding cases therefore use no bindings map.
Pattern and lexical semantics are unchanged.


## Cut 12 — Exact-size AST collection realization

Status: ACTIVE

Three remaining evaluator paths still used growth-based slices even though the
immutable AST makes the final collection size knowable before realization:

- function parameter-name extraction;
- list literal element realization;
- pipe argument construction.

Cut 12 performs a cheap AST count pass and allocates each collection once at
its exact size. Pipe evaluation preserves the existing `[]Value` call ABI and
the same left-to-right evaluation order; it only removes backing-array growth.

This is the same realization class as gated Cut 5, generalized to the remaining
known-size evaluator collections.


## Cut 13 — Resolve semantic call probe once

Status: ACTIVE

`applyFunction` previously recovered the same dynamic semantic probe separately
for the semantic call event, call-realization counters, and Number call-return
realization. Under profiling this occurs on every user and substrate call.

Cut 13 resolves the probe once at semantic call entry and threads that internal
observer reference through the two realization counters. Context-required
substrate calls still recover the active probe for `EvalContext` exactly as
before.

The probe is evaluator execution metadata rather than Aiki-visible state; this
does not alter value, scope, call, or profiling semantics.


## Cut 14 — Remove redundant call-environment metadata work

Status: ACTIVE

`NewCallEnv` already inherits the caller's dynamic semantic probe, stack, stack
limit, and the lexical function scope. `applyUserFunction` immediately
re-resolved and rewrote the same semantic probe on every user entry.

Cut 14 removes that redundant write and resolves call-site line/scope once per
entry before stack-frame enter/reuse. This does not alter lexical lookup,
dynamic profiling inheritance, authority, stack-limit, or tail-call semantics.


## Cuts 12–14 reconciliation

The first delivery of Cuts 12, 13, and 14 used separate whole-file overlays
derived from the same Cut 9–11 baseline. Because the cuts overlap in
`functions.go`, applying them sequentially could let a later overlay replace an
earlier change.

This reconciliation is the authoritative cumulative realization of Cuts
12–14. It preserves all three intended changes simultaneously. No semantic
design decision changed; this is delivery-state correction.


## Cut 15 — Proven non-escaping tail call environment reuse

Status: ACTIVE

Tail-call execution already reuses the logical stack frame but normally creates
one fresh `Env` for every tail jump. Reusing that environment indiscriminately
would be incorrect: a `func_literal` evaluated in the current invocation can
capture the call environment and survive the jump.

Cut 15 adds a conservative structural proof at function creation. A function is
eligible only when its body contains no nested `func_literal` anywhere. On a
tail jump from such a function, the existing call environment is reset and
rebound for the target invocation rather than replaced. Bodies capable of
creating closures retain the existing fresh-environment path.

`ResetCallEnv` clears invocation-local bindings, shapes, deleted-parameter state,
source/module metadata, and updates lexical/dynamic relationships before reuse.
The existing stack frame remains shared and is replaced exactly as before.

Profiling adds `tail_env_reuse` separately from logical `tail_reuse`, so the
proof's actual firing rate is observable.
