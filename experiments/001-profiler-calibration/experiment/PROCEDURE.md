# Procedure — Experiment 001: Profiler Calibration

## Question

Before trusting the very large profiler counts produced by recursive
self-hosting, does the profiler report work correctly and additively on
workloads whose structure we can understand in advance?

## Rationale

The experiment deliberately begins with programs simple enough to count by
inspection. Only after those checks pass does it introduce self-interpretation.
The progression is intended to make the scale understandable rather than ask
us to trust a large recursive measurement on sight.

## Run

Make sure the Aiki executable you want to examine is on `PATH`, then from this
directory run:

```sh
./run.sh
```

The runner records:

```sh
command -v aiki
aiki -v
```

and writes the complete transcript to a timestamped file under `../results/`
while also displaying it on standard output.

## Stage 1 — visible decade scaling

Materials:

- `sanity-10.ai`
- `sanity-100.ai`
- `sanity-1000.ai`
- `sanity-10000.ai`

Each program places the same leaf calculation:

```aiki
1 + 1
```

inside one, two, three, or four loops of ten.

The leaf calculation therefore executes:

```text
10 -> 100 -> 1,000 -> 10,000
```

The expected total completed loop-body iterations are:

| case | leaf calculations | total loop iterations |
|---|---:|---:|
| one loop | 10 | 10 |
| two loops | 100 | 110 |
| three loops | 1,000 | 1,110 |
| four loops | 10,000 | 11,110 |

This is the sanity check. If the profiler cannot reproduce the visible
structure of these programs, stop here.

## Stage 2 — exact native expression

`native-exact.ai` contains:

```aiki
1 + 2 + 3 + 4
```

Expected:

```text
arithmetic = 3
```

## Stage 3 — simple native loop scaling

Materials:

- `native-loop-10.ai`
- `native-loop-100.ai`
- `native-loop-1000.ai`

Each uses the same loop shape:

```aiki
while i < N {
    x = x + 1
    i = i + 1
}
```

Expected:

```text
arithmetic = 2N
comparison = N + 1
iteration  = N
```

## Stage 4 — one-level self-host differential

Materials:

- `selfhost-1x-0.ai`
- `selfhost-1x-1.ai`
- `selfhost-1x-2.ai`
- `selfhost-1x-4.ai`

Aiki loads the Aiki-written interpreter once and asks it to evaluate the same
tiny source expression zero, one, two, or four times.

The zero case establishes fixed cost. The question is whether each additional
identical interpreted evaluation adds the same semantic work.

## Stage 5 — bounded two-level self-host differential

Materials:

- `selfhost-2x-0.ai`
- `selfhost-2x-1.ai`
- `selfhost-2x-2.ai`

This repeats the same idea one interpreter level deeper. The zero case measures
the fixed cost of reaching that level; the one- and two-evaluation cases test
whether repeated third-level evaluations add the same semantic-count vector.

Only zero, one, and two repetitions are used because this level is deliberately
expensive.

Each run is recorded as `../results/run-YYYY-MM-DD-HHMMSS.mmm.txt`.

## Measurements

The profiler reports Aiki semantic work:

- arithmetic
- comparison
- call
- iteration
- index
- send
- receive
- store_read
- store_write

and Go substrate realization:

- elapsed
- alloc_bytes
- mallocs
- gc_cycles

`alloc_bytes` is cumulative allocation traffic during the measured run, not
resident memory.

## Expected relationships

The profiler has earned confidence for the recursive-scale experiment if:

1. the 10 -> 100 -> 1,000 -> 10,000 sanity series follows its visible source
   structure;
2. the exact expression reports three arithmetic events;
3. native loops match the exact formulas above;
4. repeated one-level self-host evaluations add stable increments;
5. repeated two-level evaluations add stable increments.

Semantic relationships are the primary calibration evidence. Elapsed time,
allocation, malloc counts, and GC behavior are runtime observations and may
vary between runs.

## Retained baseline

A retained reference observation is stored at:

```text
../results/reference-2026-08-15.txt
```

Its interpretation is stored at:

```text
../analyses/interpretation.md
```
