# Milestone 14 - mundane systems substrate

## Intent

Add the small host/system capabilities needed for serious systems work without expanding Aiki syntax.

## Added surface

### `system`

- `system.args()` returns arguments supplied after the Aiki source filename.
- `system.env(name)` returns the host environment variable value or `[@error, :environment, msg]` when absent.
- CLI argument handling explicitly excludes the interpreter and source filename rather than exposing raw `os.Args`.

### `file.list(path)`

Returns the immediate directory entry names as a lexically sorted list of strings. Sorting is intentional so results are deterministic. Failures return `[@error, :io, msg]`.

### Random-access file I/O

- `file.read_at(file, offset, count)` reads bytes at an absolute position without moving the sequential cursor.
- `file.write_at(file, offset, bytes)` writes bytes at an absolute position without moving the sequential cursor.
- `file.open(path, :read_write)` opens or creates a file for non-truncating reading and writing.

`read_at`/`write_at` were chosen instead of `seek` because the existing `read_line` facility maintains buffered sequential state; absolute-position operations avoid hidden interaction with that cursor.

## Store-sharing clarification

A store may cross spawn isolation only when passed explicitly as an argument. The Aiki documentation states the explicit-sharing rule. The Go `Store` implementation documents its per-instance `RWMutex` as an implementation mechanism for race-safe shared access, not a language feature.

## Validation

- new `system` Aiki tests: 2/2 pass;
- full Aiki suite: 408/408 pass;
- file random-access and directory-list unit tests pass;
- module help/doc/export invariants pass;
- behavior smokes: 34/34;
- grammar coverage: 32/32 productions across 10 inputs;
- engine gold checks: 10/10 inputs.

## Restart

Next ordered item: assess whether a language-services layer should be extracted now, and if so identify the smallest authoritative seam before editor or notebook adapters.
