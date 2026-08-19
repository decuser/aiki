# Native/FFI architecture mutation evidence

Baseline for mutations: clean branch after the native/FFI implementation and provider-path invariant, using a disposable copy with `go.mod` changed only from Go 1.24.0 to 1.23.0 because the sandbox cannot fetch the Go 1.24 toolchain.

Each mutation was applied only to a disposable copy and discarded afterward.

| Mutation | Expected guard | Result |
|---|---|---|
| `bytes/native` imports `string/ffi` | transitive native purity | FAIL as intended; reported `bytes/native` reaches `[math/ffi string/ffi]` |
| `string/ffi` imports `string/native` | FFI native-fallback prohibition | FAIL as intended; reported fallback to `string/native` |
| explicit bare `bits` package shadows native default | bare-default ownership | FAIL as intended; reported bare `bits` resolves to `bits`, wanted `bits/native` |
| `_bits_and` mislabeled from provider to native/runtime role | provider-backed FFI export | FAIL as intended; reported `bits/ffi.bit_and` has no provider-backed implementation path |
| `turtle` imports `math/ffi` | transitive native purity | FAIL as intended; reported `turtle` reaches `[math/ffi]` |

Focused unmutated check:

```text
go test ./engine/runtime/modules ./engine/runtime/primitives ./cmd/subcommands/tools/check
PASS
```

This evidence does not substitute for the repository Go 1.24 build or `make rigorous`; substrate compilation is blocked in the sandbox by uncached external modules.

## Witness-lock mutation

A disposable copy appended a comment to frozen `lib/bits/bits_test.ai`. `check-native-ffi-witnesses.py` failed and reported the changed Git blob ID. This verifies that the preservation lock itself detects edits to old behavioral witnesses.
