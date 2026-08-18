# Procedure — Gate 1

## Question

Can five independent Aiki OS processes implement a deterministic, barrier-synchronized, multi-population Life simulation using only Aiki process endpoints and a coordinator-owned Store?

## Method

1. Start `coordinator.ai` as the fifth Aiki process.
2. The coordinator starts four copies of `worker.ai`, one for each population.
3. For every generation, the coordinator serializes the immutable current grid as one newline-delimited structured-text frame and sends the same frame to each worker.
4. Each worker computes only the cells belonging to its population in the next generation and returns one newline-delimited proposal frame.
5. The coordinator waits for all four frames, commits the next Store-backed grid, and only then begins another generation.
6. Headless acceptance runs the same seed and fixed inputs twice and requires byte-identical committed-generation transcripts.
7. Optional canvas mode renders only committed generations.

## Protocol

Gate 1 deliberately uses one human-inspectable frame per line.

Coordinator to worker:

```text
GEN|generation|width|height|grid
```

`grid` is exactly `width*height` decimal owner digits (`0` dead, `1..4` live populations).

Worker to coordinator:

```text
DONE|owner|generation|comma-separated-live-indices
```

A packed-byte protocol is explicitly deferred. Gate 1 is a process/concurrency experiment, not a serialization experiment.

## Determinism criterion

Repeated headless runs with the same seed and fixed external inputs must produce identical committed generation sequences.

## Caveats

Terminal-window creation is outside Aiki. A host may arrange separate windows for the five processes, but the systems claim concerns independent Aiki processes and their endpoints.

## Gate 2 — Load-bearing worker domains

Gate 2 retains the Gate 1 coordinator/barrier/protocol unchanged. Each worker starts from its ordinary Life proposal and may protect one cell already owned in the immutable current generation. This makes its recurring domain activity affect the next committed generation without introducing cross-owner claims.

- A reads `data/engine-a.pattern` using file/path/string/bytes.
- B uses inherited runtime environment plus seeded random/hash; current time is only a fallback seed when no explicit seed is supplied.
- C launches the project-controlled `helper.ai`, reads its stdout endpoint, checks output with regex, and parses hexadecimal through `number.from_base`.
- D uses pure list pipelines, sorting, filtering, reduction, shapes/match, and exact math.

Gate 2 passes only if `run.sh` remains byte-for-byte deterministic across repeated headless runs with the same seed and inputs.

## Gate 3 — systems spine

Interactive/system modes add the coordinator-owned systems spine:

- exclusive lock on `../results/four-way-life.lock` and structured generation log;
- portable signal watch for interrupt/terminate;
- TCP observer listener on `127.0.0.1` (port 0 by default, or `FWL_PORT`);
- terminal detection/size reporting;
- optional raw terminal state when `FWL_RAW_TERMINAL=1`;
- ordered teardown of observers, signal subscription, terminal state, log and lock.

Run the canvas/system spine:

    aiki coordinator.ai canvas 400 60 40 42

The coordinator prints the bound observer address on stderr. In another terminal:

    aiki observer.ai PORT

Closing the canvas or delivering a watched signal stops scheduling new generations and runs the coordinator teardown path.

## Gate 3 acceptance

`gate3.sh` exercises the systems spine independently of the visual demo:

1. starts the coordinator in `systems` mode;
2. verifies the exclusive results lock is held;
3. sends the coordinator `SIGINT`;
4. requires the Aiki signal channel to observe `:interrupt`;
5. requires graceful shutdown and lock release;
6. verifies a nonempty generation log;
7. exercises terminal raw/restore when stdin is a TTY.

The interactive canvas + TCP observer remains the visual/network acceptance path.
The terminal status line uses carriage-return text only; ANSI drawing remains Aiki policy and is not required for this gate.


## Gate 4 — hardening and acceptance

Gate 4 adds no new capability. It proves the composed experiment remains
deterministic, recovers from a deliberately failed Worker C subprocess,
and retains the Gate 3 systems lifecycle guarantees.

Run:

    ./gate4.sh

The gate performs:

1. the full two-run deterministic five-process acceptance (`run.sh`);
2. a forced `helper.ai` exit with status 7 through `FWL_HELPER_FAIL=1`;
3. confirmation that Worker C treats that subprocess failure as local and the
   coordinator still completes the requested generations;
4. the full signal/lock/log/terminal systems acceptance (`gate3.sh`).

A non-TTY environment may report `TERMINAL SKIP`; an actual terminal raw or
restore failure is a hard gate failure.

After `gate4.sh` passes, repository completion still requires `make validate`
from the repository root.


## Final showcase acceptance

Run:

    ./showcase.sh

Expected presentation:

1. the current terminal is the coordinator/log terminal;
2. the coordinator opens the Life canvas;
3. four worker-view terminals open, one each for A/B/C/D;
4. the worker windows show live generation status from the actual workers;
5. the coordinator reports its TCP observer endpoint and terminal state;
6. an optional `observer.ai PORT` client may connect from another terminal;
7. closing the canvas ends the generation loop;
8. workers are stopped and reaped;
9. worker-view terminals stop and close with the coordinator.

The worker-view terminals are intentionally presentation views. Worker
stdin/stdout process pipes remain owned exclusively by the coordinator, so the
validated five-process protocol architecture is unchanged by the showcase layer.
