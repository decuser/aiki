# Aiki

Aiki is an experimental programming language designed as a learning field. Its goal is to make language behavior inspectable enough to support learning, debugging, and careful reasoning about programs.

Aiki uses a small syntax, exact rational arithmetic by default, left-to-right evaluation, explicit grouping, recoverable errors as values, and an inspectable prelude and library surface.

## Status

Aiki is currently an alpha release. The language is real, usable, and extensively validated, but syntax, libraries, tooling, and interfaces may continue to change as the design is refined.

Aiki is developed and tested on Linux. Linux is the supported target for the alpha. Prebuilt macOS and Windows binaries may also be provided, but those platforms are not currently targeted or tested. They are intended for advanced users willing to diagnose platform-specific issues; your mileage may vary.

## Alpha Expectations

Aiki is usable, but it is still alpha software. Syntax, libraries, tooling, and interfaces may change as the language is refined. Performance is not yet a primary optimization target, and some areas are intentionally incomplete rather than treated as stable commitments.

## Getting Started

Aiki is implemented in Go and uses Ebitengine for graphics. On Linux, Aiki needs Ebitengine's native system libraries even when using a prebuilt binary. Building Aiki from source additionally requires Go 1.24 or later and `make`.

On Debian or Ubuntu, install `make` and the Linux packages required by Ebitengine:

```bash
sudo apt install make gcc libc6-dev libgl1-mesa-dev \
    libxcursor-dev libxi-dev libxinerama-dev \
    libxrandr-dev libxxf86vm-dev libasound2-dev \
    pkg-config
```

If you are building from source, install Go 1.24 or later. From the Aiki repository, run:

```bash
make validate
```

This builds Aiki and runs the complete validation suite. Then start the REPL:

```bash
./aiki
```

Or run an Aiki program:

```bash
./aiki extra/samples/pipeline.ai
```

Prebuilt binaries are convenient for trying Aiki. Building from source is recommended for development, validation, and following the alpha as it evolves.

## Try Aiki

Sample programs are in `extra/samples/`. Useful starting points include:

```text
extra/samples/pipeline.ai
extra/samples/newton-rec.ai
extra/samples/newton-imp.ai
extra/samples/face.ai
extra/samples/turtle-star.ai
```

## Documentation

The main language documents are:

```text
docs/this-is-aiki-formatted.odt
docs/ra1.odt
```

Additional implementation and contributor documentation is in `docs/`, including:

```text
docs/adding-to-aiki.md
docs/debug.md
docs/decisions.md
docs/testing.md
docs/profiling.md
```

## Forthcoming

A few little books about Aiki are on the way, including material aimed particularly at non-programmers, younger learners, and readers who want to explore programming without first learning a large amount of machinery. They will complement the repository documentation and take a slower, more playful route into the language.

For example, Aiki's `turtle/simple` library keeps the first steps deliberately small:

```aiki
use("turtle/simple")

new(400)
pencolor(:green)
pen_size(3)

forward(50)
right(90)
forward(50)
right(90)
forward(50)
right(90)
forward(50)
```

## Developing Aiki

Run the full read-only validation suite with:

```bash
make validate
```

Aiki distinguishes checking, blessing, and validation:

```bash
make check      # correctness checks; never writes golds
make bless      # after a verified intentional change, update blessed golds
make validate   # check and compare against blessed golds; never writes them
```

See `docs/testing.md` before updating gold files and `docs/adding-to-aiki.md` before extending the implementation.

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
ai/                      gated engineering record
```

## Roadmap

Aiki will continue to evolve toward a smaller, clearer, more inspectable language and implementation. Work will focus broadly on hardening the language, improving its tools and documentation, extending its usefulness as a systems and teaching environment, and continuing to test whether its design commitments hold under broader use.

The project does not currently promise API, library, or syntax stability. Changes will be driven by clarity, coherence, and observable behavior rather than compatibility for its own sake.

## Feedback

Aiki is in alpha. Bug reports, reproducible failures, documentation corrections, and discussion of language behavior are welcome.

## AI and Authorship

Aiki was conceived, architected, designed, specified, and directed by Will Senn (decuser). Generative AI was used extensively as an implementation instrument under continuous human direction, supervision, and control. It produced source code, tests, tooling, documentation support, and related engineering material. It did not possess design authority.

The principal architectural commitments of Aiki—exact rational arithmetic, isolated spawn concurrency, the HAL boundary, left-to-right evaluation, shaped data, the scope-gated registry, the value model, the module system, and the executable-documentation stance—were established by Will Senn and imposed on the implementation. Generative AI was sometimes asked to suggest alternatives, formulations, tests, or approaches. Those suggestions had no standing by virtue of being generated. They became relevant only when examined, accepted, modified, or rejected by the designer.

Language semantics, architecture, constraints, requirements, and acceptance criteria remained under the designer's authority and were not derived from generated implementation. Output was evaluated solely against observable behavior, specifications, executable documentation, tests, gold files, structural invariants, and other externally imposed validation gates. Generated material was accepted, revised, rejected, or replaced according to those criteria. When it was correct, it was correct because it satisfied requirements established by the designer, not because the generated artifact was treated as authoritative.

The designer worked from the language's observable surface, specifications, and documented semantics rather than from the implementation source. This separation was intentional: the implementation was required to conform to the language, not the language to the implementation. Design authority was exercised through specifications, acceptance criteria, executable documentation, behavioral tests, gold files, structural invariants, and a gated engineering record.

Generative AI operated under a comparable separation: receive an explicit task, produce candidate material, return it for evaluation. Context could accumulate; authority did not. Continuity of conversation never became continuity of design ownership. The methods evolved as the project matured, but the authority boundary did not: implementation assistance, criticism, and suggestion never became independent design authority.

Generative AI therefore participated substantially in the engineering of Aiki. Nothing entered the language merely because an AI suggested or generated it. Adoption remained an act of design judgment. The AI produced code under Will Senn's direction, within an architecture he determined, against behaviors, constraints, and validation standards he controlled. The generated implementation realizes a design whose authority lies outside the generated code. The design is his.

## Invariant Framework

Aiki's implementation is governed by an invariant framework that keeps the executable system, the documented language, and the distributed source tree in agreement. The governing principle is simple: implementation code is replaceable; observable language behavior and declared structure are authoritative.

The framework operates at multiple levels. Behavioral tests establish expected language semantics. Gold files preserve exact observable results and detect unintended drift. Executable documentation runs documented examples against the language itself so that examples cannot silently diverge from implementation. Grammar coverage requires that declared syntax have corresponding implementation behavior. Structural checks enforce repository and architectural boundaries. Library documentation and help are checked against exported interfaces. Engine smoke tests exercise the language pipeline independently of ordinary unit tests.

These are reinforced by explicit executable couplings between parts of the system that could otherwise drift independently, including:

- grammar-to-handler coverage
- prelude-to-help consistency
- formatter-to-AST coverage
- modules-to-exports consistency
- behavior-to-gold agreement
- library help and documentation-to-export agreement
- graphics-boundary confinement
- canvas transcript golds
- documentation examples-to-stated-values agreement
- documentation-entry disposition
- module-documentation presence

Distribution structure is further constrained by `treecheck`: every shipped file must be either structurally justified or explicitly allowed.

The purpose is not merely to accumulate tests. Each invariant expresses a relationship the project intends to remain true. A change is acceptable only when the affected relationships continue to hold or are deliberately revised together. This makes drift visible: syntax cannot quietly outrun the evaluator, documentation cannot quietly cease to describe executable behavior, exports cannot silently escape their help surface, and implementation reorganization cannot arbitrarily alter the distribution.

The framework is reinforced by a gated working method. Changes are made in small serial cuts, validated at the appropriate level, and recorded with rationale, discoveries, limitations, and restart point. Claims of completion are therefore tied to evidence rather than to the production of code.

This framework is especially important given the project's extensive use of generative AI. Generated implementation is not trusted because it is plausible, idiomatic, or internally consistent. It is constrained from outside itself. The tests, gold files, executable documents, structural checks, and other invariants form the boundary between generated implementation and language authority. They are the means by which an implementation produced with substantial AI assistance remains subordinate to a human-defined language rather than gradually redefining that language through implementation accident.

## License

Aiki is licensed under the BSD 3-Clause License.

Copyright (c) 2026, William D. Senn. See `LICENSE`.
