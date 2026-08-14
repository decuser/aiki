# Aiki

Aiki is an experimental programming language designed as a learning field. Its goal is to make language behavior inspectable enough to support learning, debugging, and careful reasoning about programs.

Aiki uses a small syntax, exact rational arithmetic by default, left to right evaluation, explicit grouping, recoverable errors as values, and an inspectable prelude and library surface.

## Status

This repository is released in conjunction with the SPLASH E submission, *Aiki: Designing a Programming Language as a Learning Field*.

## Build

Requires Go 1.24 or later.

```bash
make build
````

Or build directly:

```bash
go build ./cmd/aiki
```

## Run

Start the REPL:

```bash
./aiki
```

Run a file:

```bash
./aiki path/to/file.ai
```

## Test and Validate

Run the full read-only validation suite:

```bash
make validate
```

Aiki distinguishes checking, blessing, and validation:

```bash
make check      # correctness checks; never writes golds
make bless      # after a verified intentional change, update blessed golds
make validate   # check and compare against blessed golds; never writes them
```

See `docs/testing.md` before updating gold files.

## Profiling

Aiki can expose both semantic work and the Go substrate work that realizes it:

```bash
./aiki profile program.ai
./aiki profile --cpu run.cpu.pprof program.ai
make profilesweep
```

See `docs/profiling.md`.

## Examples

Sample programs are in:

```text
extra/samples/
```

Useful starting points:

```text
extra/samples/pipeline.ai
extra/samples/newton-rec.ai
extra/samples/newton-imp.ai
extra/samples/face.ai
extra/samples/turtle-star.ai
```

## Repository Layout

```text
cmd/                     command entry points
engine/syntax/           lexer, parser, grammar, grammar help
engine/semantics/        evaluator and runtime values
engine/runtime/          HAL, prelude, runtime support
lib/                     Aiki library packages
extra/samples/           example programs
extra/editors/           editor support
test/                    tests and expected outputs
docs/                    design and contributor notes
```

## Documentation

```text
docs/adding-to-aiki.md
docs/debug.md
docs/decisions.md
docs/testing.md
docs/profiling.md
docs/ra1.odt
docs/this-is-aiki-formatted.odt
```

## License

BSD 3 Clause.

