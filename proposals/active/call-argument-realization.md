# Proposal: Call Argument Realization

## Status

**ACTIVE — measurement tranche**

## Selection evidence

After immutable source-number realization, parser/lexer startup is classified
separately from steady execution.

The strongest repeated execution-side allocation site not already owned by a
completed representation project is `evalCallArgs`.

Post-literal evidence:

```text
three-level self-host
  evalCallArgs
    flat allocation space     ~199 MB
    flat allocation objects   ~7.82 M

PDP-11 Cut 7 10x
  evalCallArgs
    flat allocation space     ~6 MB
    flat allocation objects   ~192k

Four-Way Life coordinator
  evalCallArgs
    flat allocation objects   ~131k
```

This is therefore material in self-host and PDP and independently visible in
the multiprocess application leg.

## Existing realization

Every nonzero-arity call currently performs:

```go
args := make([]value.Value, arity)
```

then evaluates each argument into that exact-sized slice.

The slice is consumed by either user-function or substrate application. Tail
calls may retain the slice in an internal tail-call sentinel across an evaluator
jump.

## Prior rejection

A previous call-allocation experiment attempted to reduce argument allocation by
changing the carrier / Env layout. Although malloc count improved, object size
and escape behavior regressed, especially in PDP. That design was reverted.

This project must not repeat that experiment.

In particular:

- do not add inline argument arrays to every Env;
- do not enlarge the hot call frame;
- do not introduce shared evaluator scratch storage before proving ownership;
- do not borrow memory across tail-call retention;
- do not alter the public substrate callable ABI casually.

## Measurement tranche

Before selecting a representation, measure call argument arity:

```text
0
1
2
3
4
5+
```

and total evaluated arguments.

The profiler reports:

```text
Call realization
  arg_arity_0
  arg_arity_1
  arg_arity_2
  arg_arity_3
  arg_arity_4
  arg_arity_5_plus
  args_evaluated
```

These counts are realization facts, not semantic units.

## Decision gate

Run the complete three-leg suite.

A new argument representation is admitted only if:

1. low arities dominate strongly enough to justify bounded specialization;
2. the design keeps temporary storage local to the call lifetime;
3. tail-call retained arguments remain independently owned;
4. substrate/user call semantics and evaluation order remain unchanged;
5. Env is not enlarged.

If the distribution is broad or the safe ownership design is not clearly
better than an exact-sized slice, close this proposal with **no representation
change**.

The working principle is intentionally conservative:

> **Do not trade a small slice allocation for larger or longer-lived runtime
> state.**

## Recursion-first ownership design

The measurement tranche is extended with a bounded implementation because call
lifetime can be proven without changing the public callable ABI.

The argument carrier is an evaluator-local pool of **exclusive external
frames**. It is capacity reuse, not value caching.

Eligibility is deliberately narrow:

```text
user Function
AND TailEnvReusable == true
AND Rest == ""
```

`TailEnvReusable` is the existing structural proof that the function body cannot
create a nested closure that captures its call environment. Rest functions are
excluded because the rest List may retain a view of the argument backing
storage. Substrate calls are excluded because callables such as spawn may retain
arguments asynchronously.

All excluded calls retain the current durable exact-sized `[]Value` path.

### Frame realization

The first candidate realization is:

```text
0 args   -> no carrier
1-4      -> inline slots in an external frame
5+       -> promote that frame to a spill []Value
```

The capacity 4 threshold is **provisional** until the arity histogram is
measured. Changing that threshold changes only realization, not ownership or
call semantics.

The frame is external to Env. No field or inline array is added to every call
environment.

### Ordinary recursion

Every simultaneously active recursive call exclusively owns its own argument
frame. A nested call cannot acquire a frame that is still active in its caller.
Frames are released only after the call environment stops borrowing their
argument view.

Thus ordinary recursion grows reusable argument storage with active recursive
depth, exactly as ordinary call lifetime requires.

### Tail recursion

Tail recursion cannot safely overwrite the outgoing invocation's argument frame
while evaluating replacement arguments. For example:

```text
f(a, b) -> tail f(b, a)
```

requires both old values while the new pair is being constructed.

The correct bounded realization is therefore a two-frame ping-pong:

```text
A owns frame 1
  evaluate replacement into frame 2
  tail jump; release frame 1

B owns frame 2
  evaluate replacement into reused frame 1
  tail jump; release frame 2
```

Tail recursion therefore uses O(1) argument-frame depth independent of tail
recursion depth, without copying outgoing parameters solely to permit reuse.

The internal `tailCallValue` carries frame ownership across the evaluator jump.
The replacement invocation becomes responsible for that frame. Before release,
the call environment explicitly drops its borrowed parameter view.

### Concurrency

The evaluator can be shared by spawned execution. There is no evaluator-global
argument depth counter or scratch array. Frames are acquired independently from
`sync.Pool`; each active invocation has exclusive ownership.

### Retention bound

A promoted frame may reuse spill capacity, but spill capacity larger than 64
arguments is discarded on release so one pathological high-arity call does not
permanently bloat the evaluator-local pool.

## Gate-1 observability

Call realization additionally reports:

```text
arg_frame_new
arg_frame_reused
arg_frame_promoted
arg_durable
arg_tail_transfer
```

Together with the arity histogram, these establish whether the frame design is
actually serving the expected population rather than merely moving
allocations.

## Gate-1 tests

The implementation adds direct invariants for:

- closure/rest exclusion;
- inline versus promoted frame realization;
- distinct frames for simultaneously active recursion;
- clearing retained Value references on release;
- a 5,000-step two-argument tail recursion with argument-frame transfer and
  reuse.

The existing TCO canaries remain authoritative for language behavior.

## Gate-1 decision

This implementation is **experimental until profiled**.

It survives only if:

1. `make validate` passes;
2. semantic and TCO behavior remain unchanged;
3. arity evidence supports the chosen compact threshold (or a nearby bounded
   threshold);
4. `evalCallArgs` allocation objects fall materially in self-host/PDP;
5. frame-new counts remain bounded relative to recursive/tail-call activity;
6. PDP and Four-Way Life elapsed time remain materially flat or improve.

If synchronization, frame objects, or promoted spill retention merely replace
the old slice cost, revert the frame realization while retaining the arity
measurement.


## Gate 1 measured arity and refinement

Corrected three-leg arity evidence:

```text
self-host
  arity 0          37
  arity 1     3126534
  arity 2     5411979
  arity 3      102165
  arity 4         249
  arity 5+      59341

PDP-11 Cut 7 10x
  arity 0          25
  arity 1       87871
  arity 2      251065
  arity 3      174475
  arity 4       23043
  arity 5+      38400

Four-Way Life coordinator
  arity 0          24
  arity 1       34948
  arity 2      146585
  arity 3       29041
  arity 4           5
  arity 5+       9600
```

Self-host calls are overwhelmingly arity one or two. The provisional four-slot
compact frame is therefore wider than the dominant workload requires.

The ownership model itself survived Gate 1:

```text
self-host
  arg_frame_new        509
  arg_frame_reused     4806909
  arg_tail_transfer     143766

PDP
  arg_frame_new          2
  arg_frame_reused     192710
  arg_tail_transfer     28214
```

This demonstrates bounded frame growth under recursion and explicit ownership
transfer under tail recursion.

### Refinement

Compact capacity changes from four to two:

```text
arity 0      no active compact slots
arity 1-2    two-slot compact frame
arity 3+     promoted spill slice
```

Release clears only the slots actually used by the logical call. Promoted
storage likewise clears only its active length. The recursion/tail ownership
model is unchanged.

Gate 1 remains open until this realization is rerun across self-host, PDP, and
Four-Way Life. The objective is to retain the allocation win without the
systematic elapsed-time tax seen in the provisional four-slot realization.
