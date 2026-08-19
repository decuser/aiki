# Experiment 004 Procedure

## Gate 1 — core diagnostics

Run from this directory with the repository Aiki executable on PATH:

```sh
aiki test diagnostics/cut1_test.ai
```

Cut 1 is gated at 42/42 tests. Cut 2 adds handbook-derived operand decoding,
all eight addressing modes, PC special forms, and the readable PDP-11 reference.

Run Cut 2:

```sh
aiki test diagnostics/cut2_addressing_test.ai
aiki test diagnostics/help_test.ai
```

Read the reference directly before the monitor is available:

```sh
aiki help.ai
aiki help.ai addressing
aiki help.ai instructions
aiki help.ai registers
aiki help.ai psw
aiki help.ai mov
```

Cut 3 completes the instruction-form workload used by Lions' `m40.s` and adds dynamic instruction/addressing-mode audit:

```sh
aiki test diagnostics/cut3_execution_test.ai
aiki test diagnostics/cut3_m40_test.ai
```

The final Gate-1 run is retained under `../results/` and is followed by the strongest applicable repository validation before the gate may be marked GATED. `MFPI`/`MTPI` previous-space translation remains owned by the later KT11-D gate; Cut 3 validates their instruction and stack-transfer boundary without claiming MMU completion.

## Current validation limitation

The initial development container does not have an Aiki executable and cannot
build the supplied Alpha-35 baseline because Go 1.24 toolchain download is
network-blocked. Static source inspection is therefore only provisional evidence; Cut 3 remains ACTIVE until its focused executable validation is performed.
