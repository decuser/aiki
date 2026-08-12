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

## Validate

```bash
make build
make smoke
make test
```

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
docs/design-principles.md
docs/adding-to-aiki.md
docs/aiki-error-handling.md
docs/debug.md
docs/line-continuation-rules.md
```

## License

BSD 3 Clause.

