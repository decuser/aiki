# Aiki Test Suite

Tests organized by purpose, not location.

## Categories

| Folder | Purpose |
|--------|---------|
| `structure/` | Verify the engine produces correct intermediate representations |
| `behavior/` | Verify outputs match expected (gold files) |
| `boundary/` | Verify isolation between layers/scopes |
| `invariant/` | Verify the machine is shaped correctly |
| `canary/` | Fail if a feature disappears |
| `contract/` | Verify interfaces are satisfied correctly |
| `property/` | Random inputs, invariants must hold |
| `regression/` | Prevent specific past bugs from recurring |
| `fuzz/` | Garbage inputs, no panics |

Unit tests remain colocated with their source files.

## Running

```bash
go test ./test/...           # all categories
go test ./test/canary/...    # single category
go test ./test/boundary/...  # single category
```
