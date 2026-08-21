# Experiment 003 — Four-Way Life

Four independent Aiki worker processes compute competing Conway-Life populations while a fifth Aiki coordinator owns the world, interprocess protocol, generation barrier, rendering, and systems effects.

The governing design and implementation gates are recorded in:

```text
../../proposals/completed/four-way-life.md
```

The experiment is intentionally multiprocess. Workers never share the coordinator's Store. Generation N is immutable while workers compute proposals for N+1; the coordinator commits only after all four replies arrive.

Gate 1 establishes the deterministic five-process Life core. Gate 2 adds the distinct load-bearing worker domains. Gate 3 adds the portable-systems spine. Gate 4 hardens failure handling and composes the acceptance paths into one final experiment gate.

Raw observations belong in `results/`; interpretation belongs in `analyses/`.

## Gate 3 systems spine

The coordinator now owns file locking/logging, portable signal handling, a TCP observer service, terminal status/optional raw mode, and graceful teardown. `observer.ai` is a minimal Aiki TCP client for watching generation summaries from another terminal.


## Final acceptance

From `experiment/`:

    ./gate4.sh

Then from the repository root:

    make validate

Interactive acceptance remains:

    aiki coordinator.ai canvas 400 60 40 42

with `aiki observer.ai PORT` from another terminal using the observer port
reported by the coordinator.


## Showcase launcher

The final interactive presentation is launched from `experiment/`:

    ./showcase.sh

The current terminal remains the real coordinator/log terminal. The coordinator
owns the canvas and the four actual worker processes.

Four additional terminal windows are presentation views of those real workers.
Each worker appends compact showcase status to its own log file while its
protocol stdout remains private to the coordinator. The view windows tail those
logs; they do not introduce a second IPC path.

The launcher uses the host terminal emulator only as presentation machinery.
On Debian-style systems where `x-terminal-emulator` resolves to
`gnome-terminal.wrapper`, the launcher uses the verified `-e` invocation form.

Closing the canvas causes the coordinator to stop scheduling generations,
shut down and reap its workers, close its systems resources, and exit. The four
worker-view terminals follow the coordinator lifetime and close afterward.
