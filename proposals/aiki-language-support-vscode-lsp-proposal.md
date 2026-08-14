# Proposal: Aiki Language Support and Language Server

## Status

Proposed.

## Summary

Provide first-class editor support for Aiki through two cooperating components:

1. a small VS Code extension for editor integration and packaging; and
2. an editor-independent Aiki language server that exposes Aiki's existing language knowledge through the Language Server Protocol (LSP).

The VS Code extension should remain thin. It should not become a second parser, formatter, help system, symbol table, or semantic authority.

Conceptually:

```text
VS Code
   |
VS Code extension
   |
  LSP
   |
aiki-lsp
   |
Aiki language services
   |
grammar / parser / linter / formatter / help / modules
```

The important artifact is not the VS Code extension itself. It is the extraction of reusable Aiki language services that can serve multiple interfaces.

A longer-term architecture is:

```text
                         +-- REPL
                         |
Aiki language services --+-- aiki-lsp ------ editors
                         |
                         +-- aiki-kernel --- Jupyter
```

The language remains authoritative. Each integration is an adapter.

---

## Goals

1. Provide useful Aiki support in VS Code.
2. Keep editor behavior consistent with actual Aiki semantics.
3. Reuse Aiki's grammar, parser, formatter, linter, help system, and module information.
4. Make semantic help available directly in the editor.
5. Avoid duplicating language rules in editor-specific code.
6. Make the main language-service layer editor-independent.
7. Support other LSP-capable editors without redesigning the system.
8. Keep language-server dependencies outside the ordinary Aiki language core where practical.
9. Ensure editor support fails when it drifts from the language rather than silently diverging.

---

## Non-goals

The first implementation should not:

- implement another Aiki parser in TypeScript;
- duplicate grammar rules in the VS Code extension;
- generate documentation mechanically from declarations;
- create a second formatter;
- invent editor-only syntax;
- make VS Code required to use Aiki;
- make the language server authoritative over the language;
- require rename, references, refactoring, semantic tokens, and every LSP feature in the first version.

The extension should be useful early and become richer incrementally.

---

## Architectural Principle

Aiki already knows the answers to most questions an editor wants to ask.

The editor integration should therefore preserve a simple authority boundary:

```text
Editor owns:
    document buffers
    cursor position
    presentation
    editor commands
    protocol requests

Aiki owns:
    syntax
    parsing
    semantics
    visible names
    modules and exports
    diagnostics
    formatting
    semantic help
    documentation
```

When an editor asks a language question, the language server should ask Aiki rather than reconstructing the answer.

---

## Components

### VS Code Extension

The extension should initially provide:

- language registration for `.ai`;
- comment and bracket configuration;
- basic indentation behavior;
- syntax highlighting;
- connection to `aiki-lsp`;
- formatting command integration;
- hover;
- completion;
- diagnostics.

The VS Code component should be deliberately small.

Where practical, purely lexical presentation features may be declarative. Semantic features belong in the language server.

### `aiki-lsp`

A separate executable should implement the Language Server Protocol:

```text
aiki-lsp
```

It should receive editor requests, translate them into Aiki language-service calls, and translate the results back into LSP structures.

It must not contain an independent model of Aiki semantics.

### Aiki Language Services

The reusable language-service layer is the architectural center of the proposal.

Conceptually, it should offer capabilities such as:

```text
Parse(source)
Diagnose(source)
Complete(source, position)
Inspect(source, position)
Definition(source, position)
Symbols(source)
Format(source)
IsComplete(source)
```

The exact API should be derived from existing code rather than imposed prematurely.

These services may later be shared by:

- the REPL;
- `aiki-lsp`;
- `aiki-kernel`;
- other interactive tools.

---

## Phase 0: Basic VS Code Recognition

Before LSP exists, a minimal extension can provide useful presentation support.

This may include:

- `.ai` file association;
- line comments using `#`;
- bracket pairs;
- auto-closing pairs;
- indentation hints;
- TextMate-style syntax highlighting;
- file icon or language identifier if desired.

This layer should remain intentionally shallow.

It is acceptable for lexical highlighting to be approximate. It is not acceptable for semantic editor behavior to become authoritative here.

---

## Diagnostics

Diagnostics should be one of the first LSP capabilities.

The language server should reuse:

- the actual Aiki lexer;
- the actual Aiki parser;
- the actual Aiki linter.

The editor should therefore report the same classes of problems as command-line Aiki tooling.

Examples include:

- lexical errors;
- syntax errors;
- unresolved names where the linter can establish them;
- invalid imports or exports where currently detectable;
- structural lint findings.

The language server should preserve Aiki source locations and messages rather than rewriting them into a parallel diagnostic vocabulary.

The desired relationship is:

```text
aiki lint <-> editor diagnostics
```

with common underlying analysis.

---

## Hover and Semantic Help

Hover is one of the strongest fits for Aiki.

The editor should expose Aiki's semantic help system rather than generating a signature from a declaration and calling that documentation.

For example, hovering over a procedure should be able to show:

```text
map

Apply a function to each element of a list and return the resulting list.

map(fn, xs)

map((x) { x * 2 }, [1, 2, 3])
```

The precise content should come from the same authored help/documentation system used by Aiki itself.

This is explicitly not autodoc.

The relationship should be:

```text
editor hover <-> Aiki semantic help
```

The editor becomes another view onto the existing explanatory system.

---

## Completion

Completion should be derived from actual Aiki-visible names.

Candidate sources include:

- language keywords;
- lexical bindings;
- prelude names;
- imported names;
- module exports;
- shapes or fields where resolvable;
- module names.

The server should not maintain a hand-written completion inventory.

Completion should respect scope.

For example, names available in one lexical or module context should not appear merely because they exist somewhere in the distribution.

The target relationship is:

```text
completion <-> Aiki name visibility
```

This may require extracting or exposing more of the parser/linter environment machinery, but that work is useful beyond VS Code.

---

## Formatting

Formatting should delegate to Aiki's canonical formatter.

The editor must not implement its own formatting rules.

The desired behavior is equivalent to:

```text
aiki fmt
```

applied to the current document or requested range, subject to whatever range-formatting support is deliberately added.

Because the formatter already reparses output and checks AST preservation, editor formatting inherits the same safety property.

The relationship should be:

```text
editor formatting <-> aiki fmt <-> AST preservation
```

---

## Go to Definition

Definition lookup is a natural second-stage capability.

It should use real Aiki binding and module information to locate:

- local bindings;
- functions;
- imported procedures;
- module exports;
- possibly shape definitions.

This should not be implemented by text search except as an explicit fallback.

The main prerequisite is a reusable representation of source-level name resolution.

That representation would also support later:

- find references;
- rename;
- semantic highlighting;
- call hierarchy.

---

## Document Symbols

The language server should be able to expose meaningful document structure such as:

- package declaration;
- top-level bindings;
- functions;
- shapes;
- exported names.

This can drive editor outlines and symbol navigation.

Again, the parser should be authoritative.

---

## Signature Help

Aiki's help metadata and authored semantic help may support signature help without inventing a new metadata source.

For a call such as:

```aiki
substring(
```

the editor may show:

```text
substring(string, start, end)
```

along with a concise semantic description where appropriate.

The displayed form should come from Aiki's existing template/help data rather than parsing source declarations heuristically.

---

## Semantic Highlighting

Basic syntax highlighting can remain lexical.

A later semantic-token implementation could distinguish things such as:

- local bindings;
- functions;
- module names;
- imported names;
- symbols;
- shape names;
- fields;
- keywords;
- literals.

Semantic highlighting should come from parsed and resolved Aiki information.

It is optional and should not precede more useful capabilities such as diagnostics, hover, completion, and formatting.

---

## References and Rename

References and rename should be deferred until Aiki has a sufficiently explicit and trustworthy source-level binding model for these operations.

A textual rename is not acceptable.

When implemented, rename should understand lexical scope, modules, exports, and shadowing.

This feature is therefore a consumer of stronger name-resolution infrastructure, not a reason to duplicate it inside the language server.

---

## Workspace and Module Awareness

A useful language server must understand Aiki modules beyond a single open document.

The server should eventually maintain a lightweight workspace model containing:

- project root;
- module search paths;
- known `.ai` files;
- package/module identities;
- exports;
- dependencies.

This should reuse the same module-resolution rules as Aiki execution.

Editor-visible import errors must agree with actual Aiki import behavior.

The relationship should be:

```text
workspace modules <-> Aiki module resolution
```

---

## Process Model

The VS Code extension launches or connects to:

```text
aiki-lsp
```

The server remains alive for the editor session and may cache:

- parsed documents;
- syntax trees;
- module metadata;
- help indexes;
- symbol information.

Caches are performance devices only.

They must never become independent semantic authorities.

When a document changes, the server updates the relevant analysis from the new source.

---

## Proposed Internal Service Boundary

The exact interface should emerge from the existing implementation, but a conceptual Go-level API might look like:

```go
type LanguageServices interface {
    Parse(filename, source string) ParseResult
    Diagnose(filename, source string) []Diagnostic
    Complete(filename, source string, pos Position) []Completion
    Inspect(filename, source string, pos Position) HelpResult
    Definition(filename, source string, pos Position) *Location
    Symbols(filename, source string) []Symbol
    Format(filename, source string) (string, error)
}
```

This is illustrative.

The important design test is:

> Could this service be used by something other than VS Code?

If the answer is no, editor concerns have probably leaked too far inward.

---

## Relationship to the REPL

The REPL and editor solve different presentation problems but ask overlapping language questions.

Possible shared facilities include:

- semantic help;
- visible-name discovery;
- parsing;
- completion;
- diagnostics;
- formatting.

The language server should therefore encourage useful extraction from REPL-specific code where appropriate.

The goal is not to force the REPL through LSP.

The goal is to prevent multiple implementations of the same language knowledge.

---

## Relationship to the Jupyter Kernel

The Jupyter kernel and language server should be siblings, not layers of one another.

```text
                      +--> aiki-lsp -----> VS Code / editors
                      |
Aiki language services
                      |
                      +--> aiki-kernel --> Jupyter
```

Shared concerns may include:

- persistent language metadata;
- parsing;
- semantic help;
- completion;
- source completeness;
- diagnostics.

Execution remains more central to the Jupyter kernel.

Static and interactive source analysis remains more central to the language server.

Neither adapter should own Aiki semantics.

---

## VS Code Commands

The extension may expose a small set of commands such as:

```text
Aiki: Format Document
Aiki: Run File
Aiki: Show Help
Aiki: Restart Language Server
```

Potential later commands include:

```text
Aiki: Run Tests
Aiki: Show Test Coverage
Aiki: Open Documentation
```

Commands should normally invoke existing Aiki capabilities rather than recreate them.

---

## Testing

### Language-service tests

Test the language-service layer independently of LSP.

Examples:

- completion respects lexical scope;
- hover returns authored semantic help;
- diagnostics match parser/linter behavior;
- formatter output matches `aiki fmt`;
- definitions resolve across imports correctly.

### LSP protocol tests

Test request/reply translation for:

- initialize;
- document open/change/close;
- diagnostics;
- hover;
- completion;
- formatting;
- shutdown.

### VS Code extension tests

Keep these focused on editor integration:

- language activation;
- server launch;
- file association;
- configuration;
- command wiring.

Do not retest Aiki semantics in TypeScript.

### Integration smoke test

Start `aiki-lsp`, open representative Aiki sources through an LSP client fixture, and verify expected diagnostics/help/completion/formatting.

This should run headlessly.

---

## Validation Couplings

The editor architecture creates several useful executable couplings:

```text
editor diagnostics <-> Aiki parser/linter
editor hover <-> Aiki semantic help
editor completion <-> Aiki-visible names
editor formatting <-> Aiki formatter
editor definitions <-> Aiki name resolution
workspace imports <-> Aiki module resolution
VS Code language syntax <-> Aiki lexical conventions
```

Where practical, these should be mechanically tested.

The extension should not be allowed to drift quietly from the language.

---

## Packaging

A likely distribution consists of:

```text
aiki
aiki-lsp
vscode-aiki/
```

The VS Code extension may either:

1. locate an installed `aiki-lsp`; or
2. bundle a compatible `aiki-lsp` binary for supported platforms.

The simpler model should be preferred initially.

The ordinary Aiki executable should not depend on VS Code or Node.js.

Likewise, the language server should not force editor dependencies into the Aiki core.

---

## Proposed Repository Shape

One possible organization is:

```text
cmd/
    aiki-lsp/
        main.go

engine/
    language/
        diagnostics.go
        completion.go
        inspect.go
        symbols.go
        definition.go

lsp/
    server.go
    protocol.go
    documents.go

extra/
    vscode/
        package.json
        language-configuration.json
        syntaxes/
        src/
```

The exact layout is secondary.

The critical boundary is:

```text
VS Code-specific code must not define Aiki semantics.
LSP-specific code must not define Aiki semantics.
```

---

## Phased Implementation

### Phase 1 — Basic editor package

Provide:

- `.ai` registration;
- comments/brackets;
- indentation;
- lexical highlighting.

This gives immediate usability with almost no architectural risk.

### Phase 2 — Language-service extraction

Establish reusable services for:

- parse;
- diagnostics;
- semantic help;
- formatting;
- visible names.

Test these directly.

### Phase 3 — Minimal `aiki-lsp`

Implement:

- initialize/shutdown;
- document synchronization;
- diagnostics;
- hover;
- completion;
- formatting.

At this point VS Code support is already substantial.

### Phase 4 — Structural navigation

Add:

- document symbols;
- go to definition;
- workspace/module understanding;
- signature help.

### Phase 5 — Rich semantic editor support

Investigate:

- semantic tokens;
- references;
- rename;
- code actions;
- inlay information where genuinely useful;
- test and coverage integration.

Each feature should justify itself in Aiki terms rather than merely because LSP supports it.

---

## Acceptance Criteria

The first language-aware release is successful when:

1. VS Code recognizes `.ai` as Aiki.
2. Aiki files have useful syntax highlighting.
3. syntax and lint errors appear as editor diagnostics.
4. hover exposes Aiki semantic help.
5. completion uses actual Aiki-visible names.
6. document formatting produces the same canonical result as `aiki fmt`.
7. the extension contains no independent semantic model of Aiki.
8. `aiki-lsp` can be used by an LSP client other than VS Code.
9. editor behavior is covered by automated tests against Aiki's own language services.

A later structural milestone adds:

10. go to definition works for local and module bindings;
11. workspace imports follow actual Aiki module rules;
12. document symbols reflect parsed Aiki structure.

---

## Open Questions

### Existing parser state

Determine whether current parser/linter APIs are suitable for repeated in-memory document analysis or whether a small language-service facade should own that adaptation.

### Incremental analysis

Full reparse on edit is probably sufficient initially for a small language.

Do not introduce incremental parsing complexity until measurement demonstrates a need.

### Name resolution

Assess how much existing linter/environment machinery can be reused for completion and definition lookup.

This is likely the main architectural work beyond protocol plumbing.

### Semantic help lookup

Determine the cleanest API for resolving the symbol under a cursor into the same help entry used by the REPL.

### Module workspace model

Define how editor workspaces identify project roots and module paths without creating editor-only import semantics.

### Distribution

Decide whether the VS Code extension expects `aiki-lsp` on `PATH` or bundles binaries.

Start with the model that introduces the least release machinery.

---

## Why This Fits Aiki

Editor support does not enlarge Aiki's semantics.

It makes existing semantics more directly inspectable while writing programs.

The fit is particularly strong because Aiki already has:

- a declarative grammar;
- a real parser;
- structural linting;
- a canonical formatter;
- authored semantic help;
- module/export knowledge;
- explicit source positions;
- a growing set of executable consistency checks.

A conventional editor extension could duplicate portions of those systems.

Aiki should instead expose them.

The desired result is simple:

> one language, one semantic authority, multiple interfaces.

VS Code is the first editor client.

It should not become part of the definition of Aiki.
