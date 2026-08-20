# Experiment 004 Procedure

## Gate 1 — core diagnostics — GATED

Run from this directory with the repository Aiki executable on PATH:

```sh
aiki test diagnostics/cut1_test.ai
aiki test diagnostics/cut2_addressing_test.ai
aiki test diagnostics/help_test.ai
aiki test diagnostics/cut3_execution_test.ai
aiki test diagnostics/cut3_m40_test.ai
```

Observed user evidence includes Cut 1 at 42/42, Cut 2 addressing at 54/54,
help at 16/16, Cut 3 execution at 73/73, and the final `m40.s` contract suite
passing. `MFPI`/`MTPI` previous-space translation remains owned by the later
KT11-D gate.

Read the PDP-11 reference directly before the monitor is available:

```sh
aiki help.ai
aiki help.ai addressing
aiki help.ai instructions
aiki help.ai registers
aiki help.ai psw
aiki help.ai mov
```

## Gate 2 — raw V6 tape bootstrap — GATED

The emulator consumes the TUHS raw `v6.tape` directly. It does not require
OpenSIMH `enblock` framing. The accepted V6 raw artifact is exactly 12,100
records of 512 bytes (6,195,200 bytes total), treated as read-only media with a
terminal tape mark after record 12,099.

If starting from the archived file:

```sh
gunzip -k v6.tape.gz
```

Run the focused synthetic/controller gate:

```sh
aiki test diagnostics/cut4_tape_test.ai
```

Then run the historical six-word bootstrap against the real raw tape:

```sh
aiki tape_bootstrap.ai /path/to/v6.tape
```

Expected terminal evidence includes:

```text
PC       100012
tape rec 000001
mem addr 001000
byte cnt 173526
word 000000 000407
word 000002 000654
TAPE BOOTSTRAP PASS
```

Observed real-media evidence: the six-word bootstrap executed against the TUHS raw tape, loaded record 0 through UNIBUS DMA, stopped in the loop at `100012`, and printed `TAPE BOOTSTRAP PASS`.

## Cut 5 — `aiki-pdp` monitor and observer windows — GATED

Run the focused noninteractive monitor/view test:

```sh
aiki test diagnostics/cut5_monitor_test.ai
```

Then launch the monitor laboratory:

```sh
./showcase.sh
```

This keeps `aiki-pdp>` in the current terminal and opens separate read-only CPU, UNIBUS, and tape windows. `AIKI_PDP_TERMINAL` selects a terminal emulator and `AIKI_PDP_PORT` overrides the default observer port 41140.

From `aiki-pdp>` attach the real tape and deposit the historical bootstrap:

```text
attach tape /path/to/v6.tape
deposit 100000 012700
deposit 100002 172526
deposit 100004 010040
deposit 100006 012740
deposit 100010 060003
deposit 100012 000777
run 100000
```

While running, CTRL-T must report status without changing machine state and CTRL-E must suspend execution and return to `aiki-pdp>`. Then:

```text
examine 000000 2
```

should show `000407` and `000654`. Observer windows must reflect the same CPU, tape, and UNIBUS state and may be closed without affecting execution.

Stop after this cut for user validation. Console/KL11 and execution of address 0 to the standalone `=` prompt belong to Gate 3.

## Current validation limitation

The development container does not have the Alpha-35 Aiki executable and cannot
build it because the repository requires Go 1.24 while toolchain/dependency
downloads are network-blocked. Static grammar/syntax evidence is therefore not
a substitute for the user-side Aiki gate.


## Cut 6 — KL11 console and standalone `=` — GATED

Run the focused console/device test:

```sh
aiki test diagnostics/cut6_console_test.ai
```

Then launch the laboratory:

```sh
./showcase.sh
```

The launcher now opens CPU, UNIBUS, tape, and KL11 observer windows. In the main terminal:

```text
attach tape 0 /path/to/v6.tape
boot tape 0
^E
run 0
```

The main terminal is the actual guest KL11 while the PDP is running. CTRL-T reports emulator status and CTRL-E suspends back to `aiki-pdp>`; every other byte, including CTRL-C and CTRL-D, is guest console input.

Gate when the real standalone image loaded from record zero prints its first `=` prompt through the emulated KL11. The KL11 observer must show transmitter activity. Stop there before RK11/RK05 work.

Observed real-media evidence: the standalone image printed `=` through KL11 after rewinding the real TUHS V6 tape. At suspension the tape showed one read / `1000` octal bytes in and record `000000`; UNIBUS NPR writes were fixed at `1000`, and the CPU was polling in the standalone command loop. Cut 6 is GATED.


## Cut 7 — RK11/RK05 and standalone `tmrk` — ACTIVE

Run the focused disk/controller gate:

```sh
aiki test diagnostics/cut7_rk_test.ai
```

Then launch the laboratory:

```sh
./showcase.sh
```

Attach the real V6 tape and a writable disk-0 pack (a missing disk file is created at exact RK05 size):

```text
attach tape 0 /path/to/v6.tape
attach disk 0 /path/to/rk0
boot tape 0
^E
run 0
```

At the standalone `=` prompt, run the distribution's own transfer program:

```text
=tmrk
disk offset
0
tape offset
100
count
1
=tmrk
disk offset
1
tape offset
101
count
3999
```

The first transfer writes the RK05 bootstrap to disk block 0. The second writes the binary filesystem to disk blocks 1..3999. Gate only when both operations complete under the real standalone program, with the RK observer showing disk-0 activity and the resulting host pack remaining exactly one RK05 cartridge in size. Stop before booting disk 0.

### Cut 7 performance diagnostic — ACTIVE, measure before optimization

The real standalone `tmrk` loader exposed a long but finite CPU-only memory-clear
loop:

```text
CLR (R0)+
CMP R0,R6
BLO loop
```

Before changing the emulator hot path, measure that exact instruction mix with
observers, tape, disk, and KL11 absent. Run the scaling sweep from the repository
root:

```sh
experiments/004-v6-emulator/experiment/diagnostics/cut7_perf_sweep.sh
```

The sweep runs independent `aiki profile --counts` processes at 1x, 2x, 10x,
50x, and 100x. One x is 256 loop iterations (768 guest instructions); 100x is
25,600 iterations (76,800 guest instructions), approximately the scale at which
the standalone pause became visibly objectionable. Each run records Aiki
semantic counts and Go realization measurements including elapsed time,
allocation bytes, mallocs, and GC cycles under `results/cut7-perf/`.

This is an evidence gate, not an optimization. Review scaling and dominant
semantic units before changing memory, UNIBUS, audit, state, or operand
representation. If source attribution is needed after the counts sweep, rerun a
representative scale with `aiki profile` (without `--counts`).
