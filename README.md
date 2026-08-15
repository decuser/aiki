# Aiki

Aiki is an experimental programming language designed as a learning field. Its goal is to make language behavior inspectable enough to support learning, debugging, and careful reasoning about programs.

Aiki uses a small syntax, exact rational arithmetic by default, left-to-right evaluation, explicit grouping, recoverable errors as values, and an inspectable prelude and library surface.

## Status

Aiki is currently an alpha release. The language is real, usable, and extensively validated, but syntax, libraries, tooling, and interfaces may continue to change as the design is refined.

Aiki is developed and tested on Linux. Linux is the supported target for the alpha. Prebuilt macOS and Windows binaries may also be provided, but those platforms are not currently targeted or tested. They are intended for advanced users willing to diagnose platform-specific issues; your mileage may vary.

## Alpha Expectations

Aiki is usable, but it is still alpha software. Syntax, libraries, tooling, and interfaces may change as the language is refined. Performance is not yet a primary optimization target, and some areas are intentionally incomplete rather than treated as stable commitments.

## Getting Started

Aiki's normal user installation is a release archive. Unpack it in a stable
location and add that directory to `PATH`:

```bash
mkdir -p ~/opt
tar xzf aiki-<version>-linux-amd64.tar.gz -C ~/opt
export PATH="$HOME/opt/aiki-<version>-linux-amd64:$PATH"
```

Put the `PATH` line in your shell startup file if you want it to persist. Then
Aiki is available like any other command:

```bash
aiki                 # start the REPL
aiki program.ai      # run a program
```

The release directory is relocatable: the `aiki` executable finds its shipped
`lib/` and `vendor/` modules relative to itself rather than relative to the
directory from which you run it. Named packages are resolved only from those
explicit distribution roots and `~/.aiki/lib`; Aiki does not recursively scan
the current working directory for packages. Local source files are imported
explicitly by relative path, for example `import("./helper", ...)`.

On Linux, Aiki uses Ebitengine for graphics and therefore needs its native system
libraries even when using a prebuilt binary. On Debian or Ubuntu:

```bash
sudo apt install gcc libc6-dev libgl1-mesa-dev \
    libxcursor-dev libxi-dev libxinerama-dev \
    libxrandr-dev libxxf86vm-dev libasound2-dev \
    pkg-config
```

### Building Aiki from source

Development additionally requires Go 1.24 or later and `make`. From the Aiki
repository:

```bash
make validate
```

Use `./aiki` when you specifically want the development build in the source
tree. To generate the user distribution and prove that it runs independently of
the source tree:

```bash
make distcheck
```

`make dist` writes both the unpacked user distribution and its `.tar.gz`
archive beside the source tree. For example, from `~/forge/dev/aiki` it produces
sibling paths named `aiki-<version>-<os>-<arch>/` and
`aiki-<version>-<os>-<arch>.tar.gz`. The user distribution contains the built
`aiki` executable, its shipped library, and the small set of files needed to
use and identify the release; it is separate from the development source tree
and does not require Go.

For a portable development/restart snapshot, use:

```bash
make baseline
```

`make baseline` writes `aiki-baseline-<version>.tar.gz` beside the source tree.
It captures the working repository including `.git` so branch, history, refs,
and AI session state are preserved, while omitting only the built top-level
`aiki` executable. This is a development baseline, not the user distribution.

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


## Experiments

Aiki distributes reproducible empirical investigations separately from the automated validation suite. New experiments are normally created out of tree while taking their sequence number from the running distribution:

```bash
cd /path/to/scratch
aiki experiment new "Profiler calibration"
```

The command reads the next number from the distribution's `experiments/` directory but creates the numbered experiment directory in the caller's current working directory. Each experiment separates procedure and materials (`experiment/`), raw observations (`results/`), and interpretation (`analyses/`). The generated runner logs each run into `results/`. Finished experiments are promoted into `experiments/` manually. See `experiments/README.md` for the experiment contract.

## Editor Support

Aiki exposes editor-independent language services through `aiki lsp`. Xed and
VS Code use thin clients over that service; nvi uses classic tags generated by
Aiki. Development installs are available from the repository:

```bash
make install-xed-plugin
make install-vscode-plugin
./aiki tags -o tags path/to/source.ai
```

The editor adapters do not implement Aiki semantics. Diagnostics, definitions,
formatting, completion, and hover come from the same language-service core.
See `extra/editors/xed/README.md` and `extra/editors/vscode/README.md` for the
editor-specific setup and live checks.

## Self-Description and Self-Hosting

Aiki includes an independent interpreter written in Aiki under `selfhost/`. It
lexes, normalizes, parses, evaluates, and self-host-loads Aiki source modules
without reusing the Go lexer/parser/evaluator implementation. Cross-implementation
invariants compare lexical, syntactic, module, and behavioral results. The final
bootstrap invariant runs the Aiki-written interpreter through itself and requires
a third-level Aiki program to produce the specified result.

The Go implementation remains the production runtime and bootstrap substrate;
self-hosting is a conformance and boundary-sufficiency proof, not a replacement
for the Go runtime.

## Repository Layout

```text
cmd/                     command entry points
engine/syntax/           lexer, parser, grammar, grammar help
engine/semantics/        evaluator and runtime values
engine/runtime/          HAL, prelude, runtime support
lib/                     Aiki library packages
extra/samples/           example programs
extra/editors/           editor support
experiments/              reproducible empirical investigations
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
