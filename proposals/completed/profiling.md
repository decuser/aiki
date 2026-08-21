## Proposal: `profile` — Semantic Work and Substrate Profiling for Aiki

> **Status (August 2026): core implemented.** Semantic counts, source attribution,
> experiments, command-line reporting, correlated Go CPU profiling, and empirical complexity classification are now
> implemented. The measurement baseline is exercised against representative
> Aiki workloads before optimization decisions are made.

### Governing idea

Aiki should expose both **what a computation does** and **what it costs to
realize that computation**. The first is an Aiki-semantic account; the second
is a substrate account. Correlation between them makes the HAL boundary
inspectable rather than presenting two unrelated profilers. Neither view is
collapsed into a synthetic total-cost number.

The implementation is described operationally in `docs/profiling.md`. The
remainder of this proposal records the original semantic-work design and the
future complexity-analysis direction.

### Purpose

Add a `profile` library that measures **Aiki-level computational work** rather than elapsed host time.

The profiler would count semantic units such as:

```text
arithmetic
comparisons
function calls
loop iterations
index operations
messages sent
messages received
store reads
store writes
```

The goal is not to reproduce a conventional CPU profiler. It is to answer questions such as:

> What work did this Aiki computation perform?

and, across several executions:

> How does that work grow as the problem size increases?

This supports ordinary profiling, algorithm study, and empirical examination of computational complexity while keeping the measured quantities tied to Aiki semantics rather than Go implementation details.

---

## Library

Proposed module:

```text
lib/profile
```

with a small public surface:

```text
profile.counts(function)
profile.experiment(size, function)
profile.complexity(results)
```

The exact names can change, but the three operations represent distinct concerns:

* measure one execution;
* conduct a controlled set of measurements;
* analyze growth across measurements.

---

## 1. `profile.counts`

```aiki
let result = profile.counts(work)
```

`work` is a zero-argument function.

Example:

```aiki
let work = () {
    sort(xs)
}

let result = profile.counts(work)
```

Semantics:

1. enable semantic counters;
2. execute `work()`;
3. disable counters;
4. return the accumulated counts.

Conceptually:

```text
[@profile_counts,
    [
        [:comparison, 4950],
        [:arithmetic, 212],
        [:call, 102],
        [:iteration, 99]
    ]
]
```

Anything performed before `profile.counts` is outside the measurement.

---

## 2. `profile.experiment`

The more general operation is:

```aiki
profile.experiment(size, function)
```

Its defining rule is:

> `profile.experiment(n, f)` measures the execution of `f(n)`.

For example:

```aiki
let make = (n) {
    make_list(n)
}

let result = profile.experiment(100, make)
```

This measures:

```aiki
make(100)
```

and associates the resulting counts with problem size `100`.

A result might have a shape such as:

```aiki
[@profile_result,
    100,
    [
        [:comparison, 0],
        [:arithmetic, 100],
        [:call, 101],
        [:iteration, 100]
    ]
]
```

### Multiple sizes

The first argument may also be a list:

```aiki
let make = (n) {
    make_list(n)
}

let results = profile.experiment([0, 10, 100, 1000], make)
```

This is defined as measuring, independently:

```aiki
make(0)
make(10)
make(100)
make(1000)
```

and returning one result for each size.

Conceptually:

```text
[
    [@profile_result, 0,    ...],
    [@profile_result, 10,   ...],
    [@profile_result, 100,  ...],
    [@profile_result, 1000, ...]
]
```

The profiler does **not** assign meaning to the size value. The programmer does that by defining `f`.

For example:

```aiki
let experiment = (n) {
    sort(make_random_list(n))
}
```

Here `n` means list length.

```aiki
let experiment = (n) {
    multiply(make_matrix(n, n), make_matrix(n, n))
}
```

Here `n` means matrix dimension.

```aiki
let experiment = (n) {
    fib(n)
}
```

Here `n` is the numeric argument itself.

The same interface therefore works with arbitrary Aiki computations.

---

## 3. Measuring an entire program

There need not be a special distinction between profiling a function and profiling a program.

Put the computation in a function:

```aiki
let program = (n) {
    # program body
}

let results = profile.experiment([10, 100, 1000], program)
```

The function boundary identifies the computation being measured.

This also gives the programmer control over setup.

For example:

```aiki
let data = load_large_dataset()

let work = (n) {
    process(first_n(data, n))
}

let results = profile.experiment([100, 1000, 10000], work)
```

Loading the dataset occurs outside the experiment; processing occurs inside it.

If setup itself is the computation of interest, simply put it inside the function.

---

## 4. `profile.complexity`

Complexity analysis consumes experiment results:

```aiki
let results = profile.experiment(
    [10, 20, 40, 80, 160],
    experiment
)

let analysis = profile.complexity(results)
```

The profiler already knows the independent values because they are stored in each `@profile_result`.

It can compare the growth of each semantic unit:

```text
size   comparison   iteration   call
10             45          10     12
20            190          20     22
40            780          40     42
80           3160          80     82
```

The analysis might report:

```text
comparison   quadratic
iteration    linear
call         linear

dominant observed growth: quadratic
consistent with Θ(n²)
```

The wording should remain explicitly empirical:

> **observed growth consistent with Θ(n²)**

rather than claiming that execution measurements constitute a mathematical proof of asymptotic complexity.

The underlying measurements remain available so the conclusion is inspectable.

---

## Semantic units

The important design question is what counts as a unit.

The profiler should count operations defined by **Aiki semantics**, not incidental Go evaluator activity.

Likely initial units include:

```text
arithmetic
comparison
call
iteration
index
```

and, as appropriate:

```text
send
receive
store-read
store-write
```

These should remain separate rather than being collapsed into a synthetic notion of total “work.”

That allows a result to say something useful about *what kind* of work is growing.

For example:

```text
comparison   Θ(n²)
call         Θ(n)
allocation   Θ(n)
```

is more informative than:

```text
work         Θ(n²)
```

---

## Instrumentation

Aiki already has an observer mechanism, but semantic profiling should not require construction or logging of a semantic event stream.

The profiler needs only **counters**.

At semantic choke points in the evaluator:

```text
operator application
function application
loop iteration
index operation
send / receive
store access
```

the engine increments enabled counters.

Conceptually:

```go
if probes&ProbeComparison != 0 {
    counters.Comparison++
}
```

There should be no event allocation, string formatting, logging, timestamps, or trace history in the profiling path.

The design principle is:

> **Count locally; analyze separately.**

---

## Runtime isolation

The expensive profiling machinery may run outside the measured execution process.

The measured Aiki process needs only:

* enabled probe state;
* semantic counters;
* a way to expose the completed counter block.

A separate profiling component can perform:

* collection;
* experiment coordination;
* comparison of runs;
* complexity analysis;
* formatting and presentation.

Thus profiling does not require the executing program to perform the analytical work it is being measured against.

A semantic trace is unnecessary. Probe correctness can instead be established using tests with known counts.

For example, a tiny program containing one comparison must produce exactly:

```text
comparison = 1
```

A ten-iteration loop can have analytically known expected counts.

Those become conformance tests for the profiling model.

---

## Proposed shapes

A minimal representation might be:

```aiki
let @profile_result [size, counts]
let @complexity [unit, growth, confidence]
```

with counts represented using ordinary Aiki data:

```aiki
[
    [:comparison, 4950],
    [:arithmetic, 212],
    [:call, 102],
    [:iteration, 99]
]
```

The representation should remain ordinary and inspectable rather than introducing a profiler-specific opaque type.

---

## Minimal first implementation

A useful first version need not solve the entire complexity-analysis problem.

It could begin with:

```aiki
profile.counts(function)

profile.experiment(size, function)
profile.experiment([sizes...], function)
```

and semantic counters for:

```text
arithmetic
comparison
call
iteration
index
```

Then:

```aiki
profile.complexity(results)
```

can be added once the measurements themselves are trustworthy.

This keeps the implementation parsimonious: establish a small, general facility for **measuring semantic work**, then build complexity analysis on top of those measurements rather than embedding complexity assumptions into the evaluator.

The central contract is especially small:

> `profile.experiment(n, f)` measures `f(n)`.
> `profile.experiment([n...], f)` measures `f(n)` independently for each supplied size.

Everything else follows from that.

