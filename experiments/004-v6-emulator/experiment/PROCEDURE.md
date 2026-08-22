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

Mnemonic examination is available for suspended-machine diagnosis:

```text
examine -m 003074 4
```

Mnemonic mode retains the octal address and instruction word, uses the CPU's
authoritative decoder for instruction classification, consumes extension words
without executing the instruction, and interprets `COUNT` as an octal number of
instructions.

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

The main terminal is the actual guest KL11 while the PDP is running. CTRL-T reports a two-line emulator snapshot (CPU state plus wall/internal time and I/O progress) and CTRL-E suspends back to `aiki-pdp>`; every other byte, including CTRL-C and CTRL-D, is guest console input.

Gate when the real standalone image loaded from record zero prints its first `=` prompt through the emulated KL11. The KL11 observer must show transmitter activity. Stop there before RK11/RK05 work.

Observed real-media evidence: the standalone image printed `=` through KL11 after rewinding the real TUHS V6 tape. At suspension the tape showed one read / `1000` octal bytes in and record `000000`; UNIBUS NPR writes were fixed at `1000`, and the CPU was polling in the standalone command loop. Cut 6 is GATED.


## Cut 7 — RK11/RK05 and standalone `tmrk` — GATED

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

The first transfer writes the RK05 bootstrap to disk block 0. The second writes the binary filesystem to disk blocks 1..3999. Gate only when both operations complete under the real standalone program, with the RK observer showing disk-0 activity and the resulting host pack remaining exactly one RK05 cartridge in size.

Observed real-media evidence: both `tmrk` transfers completed and returned to the standalone `=` prompt while the tape advanced through the binary filesystem. The resulting RK0 pack contains block 0 plus disk blocks 1..3999. Cut 7 is GATED.

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

## Cut 8 — RK05 block-zero bootstrap and V6 disk boot — ACTIVE

Run the focused disk-bootstrap gate:

```sh
aiki test diagnostics/cut7_rk_test.ai
```

The monitor command is:

```text
boot disk UNIT
```

For RK0 built by Cut 7:

```text
boot disk 0
```

The monitor deposits the RK11 bootstrap at CPU address `173110`, corresponding
to the DEC BM792-YB physical bootstrap entry `773110` on the PDP-11/40 I/O/ROM
page. The bootstrap programs RKWC for one 256-word sector, RKBA for address 0,
RKDA for block 0 on the selected unit, issues RK11 READ+GO, waits for READY, and
transfers control to address 0.

The focused gate must prove one 512-byte RK11 read into low memory and transfer
of control to address 0. The real-media gate is stronger: boot the RK0 produced
by the standalone `tmrk` procedure and require the V6 block-zero bootstrap to
reach its `@` prompt through KL11. The subsequent target is `@rkunix` and a
successful V6 startup.


## Cut 9 — UNIBUS I/O observation and KW11-L line clock

Live disk boot observation reached the block-zero `@` prompt, accepted
`rkunix`, completed 0167 RK05 reads, and then continued substantial CPU/UNIBUS
activity with RK and KL11 counters stable. Cut 9 adds two architectural pieces
before further boot diagnosis.

First, the UNIBUS records the most recent CPU-visible I/O-page DATI, DATO, and
DATOB addresses. CTRL-T includes these addresses on its second observation line
and the UNIBUS observer shows them explicitly. Observation reads existing state
only and does not itself issue UNIBUS transactions.

Second, add the KW11-L line-time clock required by the V6 `CLOCK1` path:

```text
CSR       177546 CPU / 777546 physical
vector    000100
priority  BR6
```

The host wall clock does not drive the guest. The clock advances from the
existing deterministic internal machine-time stream. In this model one nominal
line event is realized every 10,000 internal ticks; this simulation scale is an
explicit implementation constant, not a claim about PDP instruction timing.
A line event sets MONITOR and, when interrupt enable is set, requests BR6/vector
100.

Cut 9 also closes the generic CPU interrupt seam needed by that device:
interrupt requests are recognized before fetch, accepted only above the current
PSW priority, stack old PSW then old PC in the order expected by RTT, load the
new PC/PSW from the vector, acknowledge the request, and wake WAIT. While WAIT
is asserted, deterministic device time continues to advance so a clock event
can wake the processor.

Gate with `diagnostics/cut9_clock_test.ai` plus the repository validation. The
live acceptance witness remains the constructed `/tmp/rk0`: `boot disk 0`,
`@rkunix`, then observe KW11-L interrupts and subsequent V6 progress without
special-casing kernel code.


## Console milestone timing

For observational boot runs, enable host-side timing before starting the guest:

```text
timing on
attach disk 0 /tmp/rk0
boot disk 0
@rkunix
```

Timing annotations are emitted on stderr and do not alter guest machine state or
the guest stdout byte stream. The timer starts at each `demo`, `boot`, `run`, or
`continue` execution segment. It records execution start, guest input lines,
first output after input or a quiet interval, and line-leading `@`, `=`, and `#`
prompts. `timing off` restores the historically clean console presentation.

These are host wall-clock observations only; they are not PDP-11 guest time and
do not drive the deterministic simulator clock.


### WAIT and deterministic device time

Cut 9 keeps the processor logically running while a PDP-11 WAIT instruction is
active. No guest instruction is fetched and the PC does not advance, but one
unit of deterministic machine/device time advances per execution cycle. An
eligible UNIBUS interrupt is considered before the WAIT hold and clears the
waiting state on acceptance. WAIT cycles therefore advance time without
incrementing the completed-instruction count.


## Cut 10 — KT11-D memory management

Cut 10 inserts the PDP-11/40 KT11-D between CPU virtual references and the
18-bit physical UNIBUS address space. It implements Kernel/User PAR/PDR banks,
page relocation, ACF protection, expansion-direction/page-length checks, PDR
written-bit maintenance, SSR0/SSR2, previous-mode MFPI/MTPI transfers, and
processor fault entry for segmentation vector 0250 and bus-error vector 0004.
Processor status at 0177776 is now backed by the real CPU PSW rather than
ordinary I/O-page storage. Unassigned physical I/O-page references now produce
a bus-timeout fault instead of silently reading/writing backing RAM.

The bounded profiling path deliberately falls back to the authoritative machine
path after SSR0 enables KT11-D. Before management is enabled, the existing hot
path remains available. This avoids maintaining two independent MMU semantics.

First live acceptance gate:

```text
boot disk 0
@rkunix
mem = ...
```

V6 `main()` uses UISA0/UISD0 plus MFPI through `fuibyte()` as a moving window
through physical memory. The Lions laboratory is configured with 24K words of
installed core: physical bytes `000000..137777`; `140000` is the first absent
physical byte. A translated reference at that boundary must bus-timeout through
vector 4, allowing the `nofault` path to return -1 and terminate memory
inventory. The 18-bit backing address space remains larger than installed core,
and the physical I/O page remains `760000..777777`. Only after memory inventory
terminates should V6 probe the KW11-L at 0177546.


### Cut 10 bootstrap ROM and bus-timeout boundary

The BM792-YB RK bootstrap occupies an explicit emulated ROM window at physical
`773110..773150`. The monitor loads that window through an operator-only control
path; guest CPU stores do not modify it. Other unassigned physical I/O-page
addresses produce bus timeout. Installed RAM extent is separately configured;
the Lions-lab default is 24K words, so ordinary physical RAM references at or
above `140000` time out unless a separately mapped device/ROM owns the address.
This preserves disk bootstrap execution while allowing V6 physical-memory
discovery to terminate at the configured core boundary rather than at the
maximum 18-bit address-space limit.

### CUT10 debugger breakpoints

The monitor provides instruction-address breakpoints for live architectural
work:

```text
break ADDR
break list
break clear [ADDR]
```

A breakpoint is checked by the bounded executor before the target instruction
executes. Resuming skips that just-hit address once, so a breakpoint can remain
armed without immediately stopping again on the same PC. Breakpoints are host
monitor state and do not alter guest memory or registers.

The standalone `cut10_csv_cret_test.ai` exercises the V6 `csv`/`cret` calling
sequence as a closed CPU linkage cycle. It is stack-neutral in isolation; a
live `mov r4,-(sp)` reaching `nofault` must therefore be treated as evidence
that KSP was already too low before that save, not as evidence that csv itself
leaks a word per call.

### CUT10 precise watch stop

For first-corruption work, enable the diagnostic watch before starting the
monitor:

```sh
AIKI_PDP_CUT10_WATCH=1 ./showcase.sh
```

The CPU latches the successfully fetched instruction address separately from
R7. A watch report therefore pairs the correct instruction address with the
latched IR even after fetch or control transfer has changed PC.

The first Kernel SP transition from `>=130000` to `<130000` is an automatic
watchpoint stop. Bounded execution observes the transition across both the
structural hot path and the authoritative machine path. Watchpoint bookkeeping
belongs to the processor state, not module-global state, so all import paths
observe the same emulated-machine debugger state.
