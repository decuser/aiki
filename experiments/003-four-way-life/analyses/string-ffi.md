# `string/ffi` optimization experiment

Four-Way Life profiling on the showcased baseline identified protocol string
processing as a generic Aiki hot path. The pure `string.split` implementation
repeatedly calls pure `substring`, indexes runes, concatenates partial strings,
and appends list elements while decoding generation frames.

This experiment adds an explicit `string/ffi` sibling and changes only
Four-Way Life's worker/protocol imports to opt into it. Bare `string` remains the
pure/reference implementation.

The FFI boundary is deliberately coarse: one split/search/replace operation is
performed per provider call. This contrasts with the rejected bytes-grid
experiment, which crossed the provider boundary for every cell access and was
slower despite allocating fewer bytes.

Acceptance:

1. `make invariant`
2. `make validate`
3. Four-Way Life deterministic acceptance
4. rerun the hot-path profile and compare worker elapsed time, allocation,
   mallocs, call count, and parser/string attribution against the showcased
   baseline.

Do not make `string/ffi` the bare default unless parity and profiling justify a
separate decision.
