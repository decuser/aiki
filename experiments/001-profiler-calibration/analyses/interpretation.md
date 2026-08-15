# Interpretation

## Result

The profiler passes this calibration.

The experiment begins with work that can be counted directly from the source,
then increases the amount of interpretation while asking the same question at
each stage:

> If the same computation is repeated, does the profiler report the same
> additional semantic work?

The answer is yes.

The retained `../results/reference-2026-08-15.txt` is the original reference observation for this experiment. A second independent run produced identical
semantic counters for every program in the experiment. The Go realization
figures—elapsed time, cumulative allocation, mallocs, and GC cycles—varied
between runs, as expected.

## 1. Visible decade scaling

The first four programs contain no self-hosting. They place the leaf
calculation:

```aiki
1 + 1
```

inside one, two, three, and four loops of ten.

That makes the leaf calculation execute:

```text
10 -> 100 -> 1,000 -> 10,000
```

times.

The total number of completed loop-body iterations is equally visible:

| Leaf executions | Expected iterations | Observed iterations |
|---:|---:|---:|
| 10 | 10 | 10 |
| 100 | 110 | 110 |
| 1,000 | 1,110 | 1,110 |
| 10,000 | 11,110 | 11,110 |

The iteration counter is exact.

Arithmetic also follows directly from the programs. Each leaf execution
performs one addition, and every loop-body iteration increments one loop
variable. Therefore:

```text
arithmetic = leaf calculations + total loop iterations
```

The reference run reports:

| Leaf executions | Iterations | Expected arithmetic | Observed arithmetic |
|---:|---:|---:|---:|
| 10 | 10 | 20 | 20 |
| 100 | 110 | 210 | 210 |
| 1,000 | 1,110 | 2,110 | 2,110 |
| 10,000 | 11,110 | 21,110 | 21,110 |

Again, the count is exact.

The comparison counts also follow the nested-loop structure. Every loop tests
its condition once more than it executes its body. The observed sequence:

```text
11, 121, 1,221, 12,221
```

is exactly the sequence implied by the four programs.

This is the basic sanity check: the profiler counts the work that is visibly
present in ordinary Aiki source.

## 2. Exact native expression

The expression:

```aiki
1 + 2 + 3 + 4
```

contains three arithmetic operations.

The profiler reports:

```text
arithmetic = 3
```

exactly.

## 3. Native loop scaling

The native-loop programs use:

```aiki
while i < N {
    x = x + 1
    i = i + 1
}
```

Their counts are known in advance:

```text
arithmetic = 2N
comparison = N + 1
iteration  = N
```

The reference run reports:

| N | Arithmetic | Comparison | Iteration |
|---:|---:|---:|---:|
| 10 | 20 | 11 | 10 |
| 100 | 200 | 101 | 100 |
| 1,000 | 2,000 | 1,001 | 1,000 |

All three relationships are exact.

## 4. One level of self-interpretation

The next stage loads the Aiki-written interpreter once and asks it to evaluate
the same small source expression zero, one, two, and four times.

The zero case establishes the fixed cost. Relative to that baseline, one
additional interpreted evaluation contributes:

```text
arithmetic       958
comparison      1,082
call            3,255
iteration         832
index             935
```

The two-evaluation case contributes exactly twice those counts, and the
four-evaluation case contributes exactly four times those counts.

Thus repeated interpreted work is accumulated linearly and deterministically.

## 5. Two levels of self-interpretation

The final calibration repeats the same idea one interpreter level deeper.

Before any third-level probe is evaluated, merely reaching that level produces:

```text
arithmetic       689,681
comparison     1,346,671
call           5,021,109
iteration        498,039
index            793,179
```

The numbers are now large, but the test remains simple: add one identical
third-level evaluation, then add another.

From zero to one evaluation, the profiler adds:

```text
arithmetic      +353,930
comparison      +388,932
call          +2,342,216
iteration       +304,091
index           +397,156
```

From one to two evaluations, it adds:

```text
arithmetic      +353,930
comparison      +388,932
call          +2,342,216
iteration       +304,091
index           +397,156
```

The two semantic increment vectors are identical.

This is the strongest result in the calibration. The profiler exhibits the
same additive behavior at a scale of millions of semantic events that it
exhibits in the ten-operation sanity check.

## 6. Repeatability

The experiment was run twice against:

```text
aiki v0.4.0-alpha-25-gb0c9a1b-dirty
```

Every semantic counter in the second run was identical to the corresponding
counter in the first run, including all native, one-level self-host, and
two-level self-host cases.

The Go substrate realization measurements were not identical. Elapsed time,
allocation traffic, malloc count, and GC cycles varied modestly between runs.
That is expected: these quantities describe realization by the Go runtime and
the executing machine rather than the language-level semantic structure.

For this reason, the semantic counters are the calibration result. The Go
realization figures are empirical observations of cost.


## Replication runs

The reference result is intentionally retained when later runs are made.

A later run of the same experiment under the same Aiki build reproduced the
reported work counts while changing only wall-clock elapsed time. That is an
especially useful result for interpreting the profiler.

The experiment distinguishes two kinds of measurement:

1. **work accounting** — the semantic counters, and the non-time realization
   counts reported for the same execution structure;
2. **wall-clock realization** — elapsed time on the executing machine.

The replicated work counts show that the profiler is not deriving its scale
from timing. The same computation is accounted for the same way even when the
machine takes a different amount of time to realize it.

This is why later timestamped runs belong beside, rather than in place of, the
reference observation. The reference fixes one concrete observation; repeated
runs show which parts of that observation are structural and which are
contingent on execution conditions.

## Conclusion

The calibration progresses from computations whose counts can be read directly
from a few lines of Aiki to recursive self-interpretation involving millions
of observed semantic events.

At the small scale, predicted and observed semantic counts agree exactly. At
one self-host level, repeated identical evaluations add exactly proportional
work. At two self-host levels, two successive identical evaluations add
identical semantic-count vectors.

The recursive profiler is therefore behaving consistently with its intended
meaning across the range exercised here. Large recursive self-hosting counts
are not being accepted merely because they look plausible: they are produced
by the same instrumentation whose behavior is exact on small, inspectable
programs and additive under repeated interpreted work.

This does not make elapsed time, allocation, malloc, or garbage-collection
figures invariant. Those remain machine- and runtime-dependent measurements.
It does establish a concrete empirical basis for interpreting the profiler's
semantic counts in subsequent recursive self-hosting experiments.
