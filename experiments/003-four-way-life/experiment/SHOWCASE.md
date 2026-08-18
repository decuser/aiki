# Four-Way Life Showcase

## Showcase launcher

The final interactive presentation keeps the validated five-process architecture
unchanged while making the four workers visible.

Run:

    ./showcase.sh

The current terminal is the real coordinator/log terminal. The coordinator
opens the canvas and starts the same four worker processes used by the headless
acceptance path. When `FWL_SHOWCASE_DIR` is present, each worker additionally
appends a compact status line to its own showcase log. Four terminal-emulator
windows tail those logs live.

Worker protocol stdout remains private to the coordinator. The display terminals
are presentation views only; they are not a second IPC path.

Defaults:

    400 generations
    60x40 grid
    seed 42

Override with `FWL_GENERATIONS`, `FWL_WIDTH`, `FWL_HEIGHT`, `FWL_SEED`, or
`FWL_TERMINAL`.

Closing the canvas or sending interrupt to the coordinator terminates the run;
the worker-view terminals follow the coordinator lifetime and close afterward.
