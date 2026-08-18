# Four-Way Life hot-path profiling

This analysis starts from the clean showcased baseline and asks where the
remaining worker cost is realized after the active-frontier optimization.

The harness profiles Worker 1 as the representative Life worker and Worker 4
as the deliberately heavier computational worker.

For each worker it captures:

- attributed Aiki semantic sites;
- correlated Go CPU profile and label tags;
- Go allocation profile by allocated space.

The coordinator is also measured for five generations to keep parent-side work
visible, but optimization decisions should begin with worker attribution because
the four Life engines are separate Aiki processes.

Run from `experiment/`:

    ./hotpath-profile.sh

Results are written under `../analyses/hotpath/`.

Optimization rule:

> Prefer a reusable Aiki/runtime/library improvement when attribution shows a
> generic cost. Add no Life-specific accelerator unless the remaining hot path
> is demonstrably Life-specific and the pure Aiki reference implementation is
> retained.
