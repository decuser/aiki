# PDP-11/40 V6 emulator closeout

Status: **COMPLETE**

Date: 2026-08-20

## Purpose

Close the current PDP-11/40 Research UNIX V6 emulator/performance effort at a
clean engineering boundary. The emulator remains useful as an architectural
stress test and as a partially working historical reconstruction, but further
PDP-local tuning is not the next engineering priority.

The project is being closed by design decision, not because the full V6
installation path has reached an acceptable interactive speed.

## What is true now

The experiment lives under:

```text
experiments/004-v6-emulator/
```

The implemented machine has distinct CPU, UNIBUS, device, and monitor/observer
components. The working device set includes TM11/TU10 tape, RK11/RK05 disk, and
KL11 console support.

The six-word TU10 bootstrap path is demonstrated: the first 512-byte tape record
is DMA-loaded and execution reaches the expected terminal bootstrap loop. CPU,
addressing-mode, tape, monitor, console, RK, structural, fixed-domain, and
architecture-focused tests were developed and gated during the effort.

The final performance work changed both measured Aiki hot paths and the actual
interactive path:

- simple one-to-one FFI forwarding wrappers became direct first-class builtin
  bindings where the Aiki layer added no behavior;
- already-bound CPU/PSW/PC/counter state avoided redundant facade lookups;
- audit paths avoided redundant symbol-to-index conversion;
- branch conditions read only the PSW bits they require;
- branches that cannot alter PSW use an audit no-change path;
- ordinary UNIBUS RAM references take the RAM path before I/O-page mapping and
  dispatch;
- the live monitor executes bounded slices of 128 guest instructions per host
  turn rather than returning through the monitor loop after every instruction;
- KL11 transmitted output is queued so bounded slices preserve guest-output
  ordering.

The delivered live-fast-path profile reduced the 10x semantic call count from
approximately 1,040,917 calls for 7,680 guest instructions to approximately
541,586 calls for the same workload. This is a substantial reduction, but the
real `tmrk` path remains too slow for comfortable interactive use.

## Closeout decision

Stop PDP-specific performance tuning here.

The remaining slowness is now better treated as evidence about Aiki's general
execution/runtime mechanics than as a reason to continue specializing the PDP
emulator. The emulator has exposed systemic costs successfully; continuing to
trim instruction-local helpers would have diminishing value and would risk
distorting an otherwise clean CPU/UNIBUS/device architecture.

No claim is made that the complete V6 tape-to-disk installation or RK05 boot
milestones were finished. In particular, the real standalone `tmrk` path remains
performance-limited.

## Deferred system issue

`store` remains an explicit design issue for the next system-level effort.

The intended design point is isolated mutable mapped memory. If isolation means a
store has a single authority, per-access mutex locking is a runtime design bug,
not an inherent cost of `store`. The current runtime, however, permits stores to
cross `spawn`, so synchronization cannot simply be removed without first
reconciling the sharing contract.

This issue is deliberately deferred out of the PDP project. It is not to be
forgotten or "optimized around" locally in the emulator.

## Validation state

The last live-fast-path cut reported green focused semantic gates for:

```text
Cut 5 monitor         39/39
Cut 6 console         24/24
live slice            16/16
structural            27/27
fixed domain          16/16
architecture          10/10
CPU extension         16/16
RK                     47/47
```

The closeout itself changes only the AI working record. It does not make a new
integration-level correctness claim beyond the gates already recorded for the
implemented cuts.

## Next action

Begin a separate system-level project. Start from the runtime contract rather than
from the PDP emulator and examine the general costs the emulator exposed,
beginning with the `store` isolation/sharing contract and other per-operation
runtime overhead.

The PDP experiment should remain available as a regression/stress workload for
that work rather than as the place where the runtime is redesigned.
