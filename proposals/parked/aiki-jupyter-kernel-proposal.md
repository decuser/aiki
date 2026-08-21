# Proposal: Aiki Jupyter Kernel

## Status

Proposed.

## Summary

Provide Aiki as a Jupyter kernel through a small adapter process that speaks the Jupyter messaging protocol and delegates language behavior to the existing Aiki implementation.

The kernel is not a second Aiki runtime and should not become one. Parsing, evaluation, environments, faults, help, documentation, formatting, and other language behavior remain authoritative in Aiki. The kernel translates between Jupyter messages and those existing facilities.

The intended executable is:

```text
aiki-kernel
```

Conceptually:

```text
JupyterLab / Notebook
        |
 Jupyter protocol
        |
   aiki-kernel
        |
 Aiki language services
        |
 evaluator + environment
```

The value of the kernel is not merely notebook execution. Aiki already has facilities that map unusually well onto the notebook model: persistent interactive evaluation, semantic help, exact arithmetic, graphics, inspectable values, and emerging semantic measurement. Jupyter provides another interface onto those capabilities without changing the language.

---

## Goals

1. Execute Aiki notebook cells in a persistent Aiki environment.
2. Preserve ordinary Aiki semantics exactly.
3. Return Aiki values, printed output, and faults through normal Jupyter channels.
4. Expose Aiki's semantic help through Jupyter inspection.
5. Use Aiki's parser to determine whether a cell is complete.
6. Keep all Jupyter-specific machinery outside the Aiki language core.
7. Make the adapter small enough to inspect and test independently.
8. Leave room for richer Aiki-specific notebook behavior without requiring it for the initial kernel.

---

## Non-goals

The initial kernel should not:

- implement a second evaluator;
- duplicate Aiki help or documentation;
- introduce notebook-only Aiki syntax;
- change Aiki evaluation semantics;
- require changes to ordinary Aiki programs;
- add a general language-server architecture as a prerequisite;
- make Jupyter a dependency of the Aiki executable;
- make rich display or graphics support a condition of first release.

The kernel should be an interface, not a new semantic layer.

---

## Architectural Principle

The kernel must preserve a simple authority boundary:

```text
Jupyter owns:
    notebook cells
    message transport
    execution requests
    displayed notebook results

Aiki owns:
    syntax
    parsing
    evaluation
    environments
    values
    faults
    semantic help
    documentation
    formatting
    language behavior
```

When Jupyter asks a language question, the kernel should ask Aiki rather than reimplementing the answer.

---

## Process Model

`aiki-kernel` runs as a separate process launched by Jupyter using a standard kernelspec.

A typical kernelspec would invoke:

```text
aiki-kernel -f {connection_file}
```

The connection file supplies the transport information required by the Jupyter protocol.

The kernel process owns one persistent Aiki session for the lifetime of the notebook kernel. Each execution request evaluates in that session unless the user explicitly resets or restarts it.

This corresponds naturally to Aiki's existing REPL behavior.

---

## Minimal Protocol Surface

The first useful kernel needs only a small portion of the Jupyter protocol.

### `kernel_info_request`

Report:

- language name: Aiki;
- Aiki version;
- kernel implementation/version;
- Jupyter protocol version;
- file extension: `.ai`;
- MIME type;
- language metadata.

### `execute_request`

Evaluate a cell in the persistent Aiki environment.

The kernel should:

1. receive source;
2. parse it using the normal Aiki grammar and parser;
3. evaluate it using the normal evaluator;
4. capture ordinary output;
5. return the final displayed value when appropriate;
6. translate Aiki faults into Jupyter error replies.

No alternate notebook evaluator should exist.

### `is_complete_request`

Determine whether entered source constitutes a complete Aiki form.

This should be derived from the Aiki parser rather than from bracket counting or notebook-specific heuristics.

Possible replies are the normal Jupyter states:

```text
complete
incomplete
invalid
unknown
```

### `inspect_request`

Map Jupyter inspection directly onto Aiki semantic help.

This is one of the strongest reasons to provide the kernel.

Aiki help is semantic first. It can answer what a construct or procedure means, how it is used, and what syntax belongs to it. The notebook should expose that existing information rather than generating API documentation from declarations.

For example, inspection of `ceil` could present material equivalent to:

```text
ceil

Round a number upward to the nearest integer.

ceil(number)

ceil(2.3)    # 3
```

The source of this information remains Aiki's help/documentation system.

### `complete_request`

Completion is useful but not required for the first executable kernel.

When implemented, candidates should come from real Aiki-visible names:

- lexical bindings;
- prelude names;
- imported names;
- module exports;
- keywords;
- possibly module names.

Completion should not maintain a parallel symbol inventory.

### Shutdown and interrupt

The kernel must support normal Jupyter shutdown.

Interrupt behavior should be investigated separately. It should use whatever execution interruption mechanism Aiki establishes rather than adding notebook-specific evaluator semantics.

---

## Session Semantics

A notebook kernel should behave like a persistent Aiki REPL.

For example:

```aiki
let x = 10
```

followed by:

```aiki
x + 2
```

should evaluate to:

```text
12
```

Bindings established by one successfully executed cell remain available to later cells.

The semantics of failed cells should match ordinary Aiki behavior as closely as possible. The kernel should not invent rollback semantics unless Aiki itself defines them.

Restarting the Jupyter kernel creates a fresh Aiki session.

---

## Output

The initial kernel needs two output classes.

### Stream output

Aiki `print` and `println` should appear as Jupyter stream output.

The adapter should capture output at the runtime boundary rather than replacing Aiki I/O semantics.

### Evaluated values

When a cell produces a displayable final value, the kernel should publish its ordinary Aiki representation as:

```text
text/plain
```

This makes the first kernel independent of rich-display machinery.

A later extension may add richer MIME representations for particular Aiki values.

---

## Faults and Diagnostics

Aiki faults should remain Aiki faults.

The kernel should translate an uncaught Aiki fault into the Jupyter error message structure while preserving:

- Aiki error text;
- source location;
- stack information when available;
- the distinction between a language fault and kernel/protocol failure.

Protocol or transport failures must not be presented as Aiki program faults.

The objective is translation, not reinterpretation.

---

## Semantic Help

The Jupyter kernel should treat Aiki's help system as a first-class language service.

This is explicitly not autodoc.

The notebook should not inspect function declarations and manufacture reference text. Instead, Jupyter inspection should expose Aiki's intentionally authored semantic help and associated syntax.

This preserves the existing relationship among:

```text
language meaning
help
syntax
documentation
examples
exports
```

The kernel therefore becomes another consumer of the same explanatory system rather than another source of truth.

---

## Formatting

Formatting is not required to execute notebooks.

A later notebook integration could expose the existing Aiki formatter, but the kernel should not embed a second formatter or silently rewrite submitted cells.

If Jupyter or an editor requests formatting, Aiki's canonical formatter must remain authoritative.

---

## Graphics

Graphics should be considered after the text kernel works.

Aiki's current canvas boundary is promising because graphics are already separated from ordinary evaluation through an explicit process/protocol boundary.

Possible later notebook behavior includes:

- rendering a completed canvas into an image;
- publishing it with `image/png`;
- preserving ordinary external canvas behavior outside notebooks;
- allowing headless notebook execution.

The kernel should not make notebook display semantics part of the canvas core.

A first kernel may simply support textual execution and leave graphics unchanged.

---

## Semantic Measurement

The emerging Aiki semantic counters could eventually make notebooks particularly useful for computational experiments.

A notebook cell could define a function or experiment, then a later cell could inspect semantic work in units such as:

- arithmetic operations;
- comparisons;
- calls;
- iterations;
- indexing;
- sends;
- receives.

This is not required for the kernel itself.

The important architectural point is that the kernel would display results produced by Aiki's measurement facilities rather than embedding a profiler in the Jupyter adapter.

---

## Proposed Internal Interface

The existing runner is oriented toward files and one-shot execution. The kernel will probably benefit from a small reusable session abstraction rather than reaching directly into evaluator internals.

Conceptually:

```go
type Session interface {
    Execute(sourceName string, source string) Result
    IsComplete(source string) Completeness
    Inspect(name string) HelpResult
    Complete(source string, cursor int) CompletionResult
    Reset() error
}
```

This is illustrative, not a required API.

The important requirement is that the same abstraction should also be suitable for Aiki's own REPL or other interactive front ends where practical. A Jupyter-only semantic interface would be a warning sign.

A session would own:

- loaded grammar;
- runtime;
- prelude environment;
- persistent user environment;
- module registry/session state as appropriate.

This avoids rebuilding the complete language environment for every cell.

---

## Implementation Boundary

A possible repository layout is:

```text
cmd/
    aiki-kernel/
        main.go

engine/
    session/
        session.go

jupyter/
    connection.go
    messages.go
    kernel.go
    transport.go

extra/
    jupyter/
        kernel.json
        install.sh
```

The exact layout is secondary to the boundary:

```text
Jupyter protocol code must not enter the evaluator.
Evaluator semantics must not enter the Jupyter protocol package.
```

If a small general-purpose interactive-session package is extracted, both the REPL and kernel may use it.

---

## Transport Dependency

The Jupyter messaging protocol uses ZeroMQ.

Before implementation, choose deliberately between:

1. a maintained Go ZeroMQ library;
2. a small supported binding around an installed ZeroMQ implementation;
3. another standards-conforming transport implementation if one is demonstrably simpler.

This dependency should be isolated inside the kernel adapter.

The ordinary `aiki` build should not acquire a mandatory Jupyter/ZeroMQ dependency merely because a kernel exists.

A separate executable and build target are therefore preferable.

---

## Kernelspec

A minimal installed kernelspec will require a `kernel.json` similar in role to:

```json
{
  "argv": [
    "/path/to/aiki-kernel",
    "-f",
    "{connection_file}"
  ],
  "display_name": "Aiki",
  "language": "aiki"
}
```

Installation should eventually be scriptable, for example:

```text
make jupyter-install
```

or:

```text
aiki-kernel install
```

The exact installation command can be decided after the executable exists.

---

## Testing

The kernel should be tested at several boundaries.

### Protocol unit tests

Test:

- message parsing;
- message construction;
- execution counts;
- request/reply correlation;
- error translation;
- kernelspec behavior.

### Session tests

Test persistent Aiki semantics independently of Jupyter:

```text
cell 1: let x = 10
cell 2: x + 2
result: 12
```

Also test:

- syntax errors;
- runtime faults;
- state persistence;
- reset;
- imported modules;
- printed output;
- final expression values.

### Help tests

Verify that `inspect_request` uses Aiki help as authority.

The test should fail if the kernel begins maintaining its own explanatory inventory.

### Completeness tests

Use known complete, incomplete, and invalid Aiki forms to verify that notebook completeness agrees with the Aiki parser.

### Integration smoke test

Launch `aiki-kernel` with a generated connection file, execute a small sequence of protocol requests, and confirm expected replies.

This should be automatable and headless.

---

## Validation Couplings

The kernel introduces several useful new couplings that can be enforced mechanically:

```text
kernel inspection <-> Aiki semantic help
kernel completion <-> Aiki-visible names
kernel completeness <-> Aiki parser
kernel execution <-> Aiki evaluator
kernel displayed values <-> Aiki value representation
kernel faults <-> Aiki diagnostics
kernelspec version <-> built executable
```

These should be tests, not conventions.

---

## Phased Implementation

### Phase 1 - Session extraction

Establish a persistent interactive Aiki session that can:

- execute source strings;
- retain bindings;
- return final values;
- capture output;
- return faults.

Prove it without Jupyter.

### Phase 2 - Minimal kernel

Implement:

- connection-file loading;
- Jupyter transport;
- `kernel_info_request`;
- `execute_request`;
- shutdown;
- stream output;
- plain-text result output;
- error replies.

At this point Aiki should execute normally in a notebook.

### Phase 3 - Language-aware notebook behavior

Add:

- `is_complete_request`;
- `inspect_request` backed by semantic help;
- `complete_request` backed by actual Aiki-visible names.

This is where the kernel becomes distinctly Aiki rather than merely executable.

### Phase 4 - Rich integration

Investigate:

- canvas/image display;
- semantic profiling results;
- richer MIME representations;
- source formatting;
- notebook-oriented teaching/research examples.

Each should remain optional and should reuse existing Aiki machinery.

---

## Acceptance Criteria

The initial kernel is successful when all of the following are true:

1. Jupyter can discover and launch an Aiki kernel.
2. A notebook can execute ordinary Aiki source.
3. bindings persist across cells;
4. printed output appears normally;
5. the final Aiki value is displayed as text;
6. Aiki faults appear as notebook errors with useful diagnostics;
7. restarting the kernel creates a fresh Aiki environment;
8. kernel execution agrees with ordinary Aiki execution;
9. the ordinary Aiki build does not require the Jupyter transport dependency.

The language-aware milestone adds:

10. cell completeness comes from the Aiki parser;
11. notebook inspection exposes Aiki semantic help;
12. completion candidates come from the actual Aiki environment and module system.

---

## Open Questions

### Interrupts

How should a running Aiki computation be interrupted safely?

This should be solved at an Aiki execution/session boundary rather than as a Jupyter-specific evaluator hack.

### Output capture

What is the cleanest runtime boundary for directing `print` and `println` to the kernel without changing their ordinary semantics?

### Module/session lifetime

Confirm which module registry and cache state should persist for the life of a notebook session.

### Graphics

Determine the smallest clean bridge from the existing canvas process/protocol to a notebook MIME result.

### Packaging

Decide whether the kernel is:

- built by default;
- an optional build target;
- distributed as a separate binary;
- or maintained as a closely coupled companion component.

The presumption should be that Jupyter dependencies do not burden the ordinary Aiki executable.

---

## Why This Fits Aiki

A Jupyter kernel does not require Aiki to become notebook-oriented.

It provides another inspectable interface to facilities Aiki already has.

The fit is particularly strong because the notebook interaction model aligns with existing Aiki properties:

- persistent interactive evaluation;
- small, direct expressions;
- exact arithmetic;
- semantic help;
- inspectable values;
- explicit faults;
- graphics;
- computational experiments;
- semantic measurement.

Most importantly, the kernel can remain subordinate to the language.

Jupyter supplies the notebook.

Aiki supplies the meaning.
