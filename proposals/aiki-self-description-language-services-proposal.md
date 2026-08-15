# Proposal: Aiki Self-Description, Conformance, and Language Services

## Status

Active — Phase I GATED; Phase II through Cut II.4 GATED; Cut II.5 at nvi handoff.

## Baseline

Implementation began from Aiki baseline `v0.4.0-alpha-14-g9c78646` (`9c78646`).

It replaces the repository's earlier VS-Code-centered language-server proposal and incorporates the separately developed self-hosting interpreter design into one implementation sequence. The substantive goals of both efforts are retained, but the work is reordered around a shared dependency: Aiki first needs an independently expressed, observable account of its own front end. That account then becomes evidence and conformance infrastructure for language services, while the full self-hosted evaluator follows afterward.

---

## Thesis

Aiki should be able to **describe itself independently** and **expose what it knows through stable services**.

These are two expressions of the same architectural commitment:

1. the grammar, documented semantics, runtime vocabulary, and HAL boundary are authorities;
2. implementations realize those authorities but do not replace them;
3. observers, probes, tools, protocols, and editors consume stable projections rather than inventing parallel language knowledge.

The combined architecture is therefore:

```text
                         AUTHORITIES

             grammar / semantics / vocabulary
                         HAL boundary
                             |
             +---------------+---------------+
             |                               |
             v                               v
      Go implementation               Aiki implementation
       lexer / parser                  lexer / parser
       evaluator                       evaluator
             |                               |
             +---------------+---------------+
                             |
                    conformance surfaces
                 tokens / syntax / behavior
                             |
                  language-service contract
                      /       |       \
                observer    probe    results
                             |
             +---------------+----------------+
             |               |                |
             v               v                v
          LSP adapter      tags adapter     CLI adapter
             |               |                |
          Xed/VS Code        nvi          lint/fmt/etc.
```

The critical distinction is:

> **Duplicate implementation is necessary for proof. Duplicate authority is not.**

The Go and Aiki lexers/parsers must be genuinely separate implementations. Otherwise self-description and conformance are weakened. But keyword sets, operator sets, grammar productions, newline policy, diagnostics, runtime vocabulary, and test expectations must remain tied to authoritative Aiki sources through executable couplings and reviewed conformance artifacts.

---

## Why Combine the Work

The self-hosting and language-services efforts overlap at exactly the point where Aiki needs stable, observable representations of its own syntax.

A self-hosted front end forces Aiki to state, independently of Go implementation structures:

- what a token observably is;
- how newline normalization behaves;
- what syntactic structure is significant;
- how positions are represented for comparison;
- what constitutes equivalent lexer/parser behavior.

Those same artifacts are needed by language services. An LSP implementation, formatter projection, editor diagnostic, tags generator, or syntax declaration should not reach arbitrarily into Go lexer/parser internals. They need stable Aiki-facing projections.

Producing those projections as a side effect of language-service extraction risks canonizing Go implementation accidents. Producing them while building an independent Aiki lexer/parser forces the representation to be neutral enough for two implementations to share.

The work is therefore deliberately sequenced:

```text
Phase I   independent Aiki front end + conformance corpus
                    |
                    v
Phase II  language services + adapters
                    |
                    v
Phase III complete self-hosted evaluator + bootstrap
```

This minimizes rework while preserving both goals.

---

## Existing Architectural Precedents

The baseline already contains the patterns this work should extend rather than bypass:

- `engine/runtime/hal.RuntimeContract` separates evaluator behavior from the Go substrate;
- `engine.Observer` observes lexing, parsing, evaluation, effects, formatting, and optionally diagnostics;
- `engine/observe.SemanticProbe` provides neutral instrumentation without forcing lower layers to depend upward;
- grammar analysis and executable couplings enforce the rule that shared language facts have one authority;
- formatter and linter logic already reuse real Aiki lexer/parser/grammar/runtime knowledge;
- `extra/editors/xed/aiki.lang` provides GtkSourceView support but demonstrates the cost of manually duplicated vocabulary: it is stale;
- `aiki fmt`, `aiki lint`, `aiki profile`, and `aiki treecheck` establish the existing command model into which `aiki lsp` and possibly `aiki tags` naturally fit.

The new subsystem should look and feel like these existing parts of Aiki rather than forming an IDE-specific island.

---

# Design Principles

## 1. Authorities are above implementations

The primary authorities remain:

- `grammar.ebnfx` for syntax structure and grammar-declared policy;
- documented semantics and executable behavior/gold artifacts for language behavior;
- runtime/prelude/module registries for visible vocabulary;
- HAL contract for the platform boundary.

Neither the Go implementation nor the Aiki implementation becomes the language simply by existing.

Agreement between implementations increases confidence; disagreement exposes a bug, ambiguity, or incomplete specification.

```text
             grammar / spec
                /     \
               /       \
             Go        Aiki
               \       /
                \     /
               conformance
```

Not:

```text
Go == Aiki
therefore correct
```

Conformance expectations must be reviewed against the authorities.

## 2. Duplicate implementation; do not duplicate authority

The self-hosted lexer and parser are intentionally separate implementations.

Necessary duplication includes:

```text
lexing algorithm
newline-normalization algorithm
parsing algorithm
eventually evaluation algorithm
```

Unnecessary authority duplication includes unchecked copies of:

```text
keywords
operators
delimiters
newline policy
grammar productions
diagnostic templates
runtime/prelude vocabulary
```

Where an independent implementation must necessarily encode such information, executable couplings should detect drift against the authority.

## 3. Conformance artifacts belong to Aiki

Token, syntax, and behavioral fixtures are not implementation-private tests.

They are Aiki artifacts consumed by multiple implementations and services.

Conceptually:

```text
test/conformance/
  syntax/
    lex/
    newline/
    parse/
  behavior/
```

The exact repository shape is a cut-level decision, but ownership is not: the corpus belongs to the language project, not to either interpreter.

## 4. Stable projections are not implementation structs

The conformance surface must not simply serialize Go lexer/parser structures and require Aiki to mimic them.

Instead, define the smallest observable forms needed for cross-implementation comparison.

For example:

```text
Token
  kind
  lexeme
  line
  column

SyntaxNode
  production/kind
  terminal value where applicable
  children
  position
```

These examples are illustrative. The actual forms should emerge during Phase I from what both implementations can naturally and meaningfully produce.

## 5. Language services are primary; protocols are adapters

LSP is a protocol projection, not the language-service architecture.

The reusable capability is the service contract:

```text
Diagnostics
Symbols
Definition
Completion
Hover/Inspect
Format
```

Adapters translate external mechanisms into that contract:

```text
LSP        -> Xed / VS Code / other LSP clients
ctags      -> nvi
CLI        -> aiki lint / aiki fmt / future queries
```

## 6. Contract, observer, instrumentation, adapter remain distinct

Use the same architectural distinctions already present elsewhere in Aiki:

```text
contract         what a consumer may ask for
observer         what happened
instrumentation  measurement of what happened
adapter          how an external system speaks the contract
```

Editors consume results. They are not observers.

Instrumentation measures service behavior. It does not define semantics.

## 7. Adapters are replaceable

An adapter is plugin-like in architecture: narrow, removable, and dependent inward on a stable contract.

This does not require Go dynamic plugin loading.

A new editor or tool should be addable without modifying grammar, parser, evaluator, formatter, or another adapter.

---

# Phase I — Independent Aiki Front End and Conformance

## Goal

Implement enough of Aiki in Aiki to independently lex and parse Aiki source, and use that implementation to establish reusable lexical and syntactic conformance surfaces.

This phase intentionally stops before evaluation.

The result is not yet a self-hosted interpreter. It is the evidence-producing front end on which both the language-services work and the later full self-hosting proof depend.

## Architectural position

The Aiki front end is an ordinary Aiki program running on the existing Go interpreter.

```text
Go substrate
  -> Go interpreter
      -> Aiki lexer/parser
          -> Aiki source being analyzed
```

It uses only prelude-level and ordinary library capabilities. No new HAL primitive or Go-specific escape hatch is introduced for the work.

### Required capabilities

Character-level lexing requires existing Aiki string/list/control capabilities such as:

```text
length
string indexing
ord / chr
first / rest / append / prepend / empty
comparison and boolean operators
if / while / match
functions and return
```

File loading may use existing library facilities where required by the harness, but the core lexer/parser should operate on source strings so conformance tests remain simple and deterministic.

---

## Cut I.0 — Grammar token authority extraction

Before writing the independent lexer, build a small Aiki program that reads `grammar.ebnfx` and extracts only the lexical authority needed for comparison.

The program is deliberately **not** a general EBNF parser. It should parse just enough of the grammar file to locate and read the `@tokens` block and normalize the declarations relevant to lexing, including as applicable:

- keywords;
- operators;
- delimiters;
- token names or categories needed by the self-hosted scanner.

Its first purpose is comparison, not generation. The independent Aiki lexer may encode its own tables, but those tables must be checkable against the grammar authority. This preserves independent implementation while preventing silent vocabulary drift.

Conceptually:

```text
grammar.ebnfx
     |
     v
Aiki token-authority extractor
     |
     v
normalized lexical facts
     |
     +--> compare with Aiki lexer tables
     +--> later compare/check editor declarations
```

This cut also provides an early exercise of ordinary Aiki file and string processing without introducing any new builtin or HAL dependency.

### Gate

An Aiki program can read `grammar.ebnfx`, extract the intended token declarations, and emit a deterministic normalized representation suitable for executable comparison.

The extractor itself introduces no second lexical authority: its result is derived directly from the grammar file.

---

## Cut I.1 — Authority and representation inventory

Inspect the exact baseline authorities and define the comparison problem with the Cut I.0 extractor now available as evidence.

Record:

- grammar source and grammar metadata relevant to lexing/parsing;
- normalized lexical facts obtainable from the `@tokens` block;
- Go lexer token kinds and observable fields;
- newline-normalization behavior and policy source;
- Go parser AST structures and which parts are semantic versus implementation-specific;
- current syntax/grammar invariant tests;
- position conventions;
- existing parser/lexer debugging projections, if any.

### Gate

A written inventory identifies:

1. which language facts are authoritative;
2. which facts the Aiki implementation must independently realize;
3. which duplicated constants require executable coupling;
4. candidate neutral token and syntax projections;
5. which lexical facts can already be checked by the Cut I.0 extractor.

---

## Cut I.2 — Aiki data model and lexical scanner

Define Aiki-native structures sufficient for lexical work. A likely token representation is:

```aiki
let @token [kind, lexeme, line, col]
```

Implement a character-level scanner in Aiki. It must independently recognize the lexical forms defined by Aiki, including whitespace/comments, names, literals, symbols, shape names, operators, and delimiters.

No regex dependency is required.

### Conformance artifact

Establish a reusable normalized token projection and corpus of known inputs/expected outputs.

Both front ends must be able to produce the same normalized representation without exposing their private token structures.

### Gate

For the Phase-I lexical corpus:

```text
reviewed expected token stream
      == Go normalized token stream
      == Aiki normalized token stream
```

Disagreements are recorded as implementation bugs or specification discoveries, not papered over in the normalizer.

---

## Cut I.3 — Newline normalization

Implement the grammar-declared newline policy independently in Aiki.

This includes suppression depth and newline-to-semicolon behavior as actually defined by the Aiki grammar/policy.

### Conformance artifact

Create explicit fixtures for newline-sensitive cases rather than relying only on downstream parser success.

### Gate

For every fixture:

```text
reviewed normalized token expectation
      == Go normalized stream
      == Aiki normalized stream
```

Any duplicated newline-policy constants in the Aiki implementation are executable-coupled to the grammar authority.

---

## Cut I.4 — Aiki parser and normalized syntax projection

### Implementation decision

The neutral syntax projection reuses the repository's existing human-readable
engine parse surface (`test/structure/engine/*.ai.parse.gold`): indentation,
grammar/token node kind, one-based line/column, and terminal lexeme where one
exists. This is grammar-shaped observable output rather than serialization of
Go structs. Reusing the existing reviewed corpus avoids creating a second parse
gold authority; the Go engine gates and the independent Aiki parser both have to
agree with the same artifacts.


Implement a recursive-descent Aiki parser matching the authoritative grammar.

A hand-written Aiki parser is intentional. It is a second implementation, not a generated wrapper around the Go parser.

A likely Aiki AST representation is shape-based, for example:

```aiki
let @node [kind, value, children, line, col]
```

but the implementation representation and conformance projection need not be identical.

### Critical design rule

Do **not** define conformance as “the Aiki AST serializes like Go structs.”

Define a normalized syntactic projection based on language-significant structure:

```text
production/kind
children
terminal kind/value where applicable
normalized source position
```

The process of making both parsers project into this representation is expected to reveal which parts of the Go AST are language concepts and which are implementation conveniences.

### Gate

For the Phase-I parse corpus:

```text
reviewed syntax expectation
      == Go normalized syntax projection
      == Aiki normalized syntax projection
```

The corpus must cover all grammar productions used by the project's grammar-coverage invariant, plus targeted malformed inputs sufficient to expose parser boundary behavior.

---

## Phase I deliverable

Phase I is complete when Aiki can independently lex, normalize, and parse Aiki source and the repository contains reusable conformance surfaces for all three stages.

The deliverable must include:

- Aiki token-authority extractor for `grammar.ebnfx` `@tokens`;
- Aiki lexer;
- Aiki newline normalizer;
- Aiki parser;
- lexical conformance corpus;
- newline conformance corpus;
- normalized syntax corpus;
- executable cross-implementation comparison harness;
- drift checks for duplicated grammar facts;
- recorded specification discoveries and resolved ambiguities.

At this point the project deliberately pauses the self-hosted interpreter before evaluator work.

---

# Phase II — Language Services and Replaceable Adapters

## Goal

Build an editor-independent language-service subsystem using the stable observation surfaces and conformance evidence established in Phase I.

The service must reuse Aiki's existing authorities and implementation machinery rather than implementing syntax or semantics again.

## Dependency direction

```text
external client
      |
      v
   adapter
      |
      v
language-service contract
      |
      v
service implementation
      |
      +--> syntax projections / lexer / parser
      +--> structural analysis
      +--> formatter
      +--> help / modules / runtime vocabulary as needed
```

Never:

```text
language core -> LSP
language core -> Xed
language core -> VS Code
language core -> nvi
```

and never:

```text
Xed parser
VS Code scope model
LSP builtin authority
nvi-specific definition analysis
```

---

## Cut II.0 — Language-service authority/dependency inventory

Revisit the baseline after Phase I and identify reusable seams for:

- source/document identity;
- token/syntax projection;
- parser diagnostics;
- lint structural analysis;
- formatter source transformation;
- runtime/prelude vocabulary;
- help/module information;
- symbol and scope analysis.

The inventory must explicitly reference the conformance projections from Phase I rather than inventing parallel representations.

### Gate

A dependency map establishes where reusable language knowledge currently resides and identifies the smallest acyclic extraction required for the first service capability.

---

## Cut II.1 — Service contract, documents, and diagnostics

Establish the minimal editor-independent contract.

The first concrete consumer is diagnostics, so do not design the entire eventual API in advance.

A document abstraction should carry at least:

```text
stable identity
source text
Aiki filename/path identity where applicable
generation/version where supplied by a client
```

The first diagnostic capability combines, as appropriate:

- lexical errors from the actual lexer;
- parse errors from the actual parser;
- structural findings from reusable lint analysis.

Results use Aiki-native positions and diagnostic types.

### Contract invariant

```text
same source authority
same location
same finding
same category/severity where applicable
```

across CLI and language-service projections.

### Gate

For shared fixtures:

```text
CLI diagnostic
    == language-service diagnostic
```

semantically, without requiring byte-identical presentation.

---

## Cut II.2 — Observation and instrumentation seam

Establish language-service observation and measurement using the same dependency discipline as existing `engine.Observer` and `engine/observe.SemanticProbe`.

The implementation must decide whether service events extend an existing neutral observer family or use a sibling contract.

Possible observable events include:

```text
document_opened
document_changed
parse_requested
diagnostic_produced
symbol_query
definition_query
completion_query
format_query
cache_hit/cache_miss
```

Possible probe measurements include:

```text
parse requests
reparse count
analysis runs
diagnostics produced
symbols visited
completion candidates
definition resolutions
format requests
```

The exact vocabulary must emerge from actual choke points.

### Gate

Observer/probe attachment is optional and behavior-neutral: enabling or disabling it does not alter service results.

---

## Cut II.3 — `aiki lsp` protocol shell

Implement the first replaceable adapter as:

```text
aiki lsp
```

Do not introduce a separate `aiki-lsp` product unless later evidence justifies it.

Initial protocol scope:

```text
initialize
initialized
shutdown
exit
textDocument/didOpen
textDocument/didChange
textDocument/didClose
publishDiagnostics
```

LSP-specific URIs, position encodings, lifecycle, and capability negotiation stay inside the adapter.

No LSP type leaks into the language-service contract.

### Gate

Opening or changing malformed Aiki source through an LSP client produces the same underlying Aiki diagnostics as the CLI/service path.

---

## Cut II.4 — Xed support and lexical declaration coupling

Xed is a first-class client because it is an actual Aiki development editor.

The baseline already contains:

```text
extra/editors/xed/aiki.lang
```

but its language inventory is stale.

Support Xed in two layers:

```text
Xed
  +-- GtkSourceView aiki.lang   lexical presentation
  +-- thin LSP integration     semantic services
          |
          v
       aiki lsp
```

Determine the smallest practical Xed-side LSP mechanism in the target environment. If generic support exists, provide configuration only. If an Aiki-specific Xed plugin is necessary, it remains a thin protocol client.

Any syntax inventory duplicated in `aiki.lang` must be generated or executable-coupled to Aiki's lexical/grammar authority.

### Gate

- `.ai` recognition and lexical highlighting are current with the language;
- stale declarations fail an invariant rather than silently drifting;
- live parser/lint diagnostics are visible in Xed through the LSP path.

---

## Cut II.5 — Symbol/definition service and nvi tags adapter

Extract reusable symbol/definition analysis from actual Aiki scope rules.

Then project it through at least two consumers:

```text
language service -> LSP definition/symbols
language service -> tags adapter -> nvi
```

nvi remains nvi. No attempt is made to embed an LSP client or recreate an IDE inside it.

A likely Unix-facing command is:

```text
aiki tags [path...]
```

producing ctags-compatible output.

### Gate

The LSP definition result and tags entry for the same definition derive from the same symbol authority and resolve to the same source location.

---

## Cut II.6 — Formatting service

Expose the existing canonical formatter through the language-service contract rather than reimplementing formatting for LSP.

The baseline formatter's parse-preservation safety gate remains authoritative.

```text
aiki fmt
      \
       -> same formatter capability
      /
LSP formatting
```

### Gate

For shared fixtures:

```text
aiki fmt source == language-service Format(source)
```

and formatting continues to reject transformations that violate its existing parse-preservation invariant.

---

## Cut II.7 — Completion and hover/inspect

Add richer semantic services only after diagnostics, definitions, and formatting have established the contract.

Completion must derive from actual Aiki visibility and scope rules.

Hover/inspect should project authored Aiki help and known semantic information rather than create a competing documentation database.

### Gate

Completion, hover, and help results are derived from existing language/runtime authorities and contain no adapter-maintained semantic inventory.

---

## Cut II.8 — VS Code client

VS Code is a thin conventional LSP client and packaging target.

Its extension may own:

- `.ai` registration;
- editor brackets/comments/indent configuration;
- lexical highlighting where useful;
- launch/connect configuration for `aiki lsp`;
- packaging.

It must not own Aiki semantic analysis.

### Gate

VS Code obtains its semantic behavior entirely from `aiki lsp` and contains no second parser, scope model, formatter, or builtin authority.

---

## Phase II deliverable

Phase II is complete when:

- an editor-independent service contract exists;
- diagnostics, symbol/definition, formatting, and selected help/completion capabilities reuse authoritative Aiki machinery;
- service work is observable and instrumentable without behavior changes;
- `aiki lsp` is a replaceable adapter;
- Xed receives live semantic support;
- nvi receives native tags/lint/fmt support;
- VS Code can operate as a thin LSP client;
- editor declarations and adapter outputs are executable-coupled to language authorities where duplication is unavoidable.

At this point Aiki has both an independent front-end conformance implementation and a reusable language-service substrate.

---

# Phase III — Complete Self-Hosted Interpreter

## Goal

Resume the independent Aiki implementation above the parser and complete the stronger self-hosting claim:

> An Aiki program, running on the Go bootstrap interpreter, can evaluate Aiki programs using only ordinary Aiki/prelude-level capabilities and can ultimately interpret itself interpreting a test program.

The Go interpreter remains the bootstrap, reference implementation, and production runtime.

This phase is intentionally downstream of Phase II because the evaluator does not materially unblock editor services, while Phase I already supplied the conformance artifacts Phase II needed.

---

## Runtime model

The self-hosted interpreter remains an ordinary Aiki program at the library/application level.

```text
Go substrate
  -> Go interpreter (bootstrap)
      -> Aiki interpreter
          -> user Aiki program
```

The HAL remains the platform boundary. The self-hosted interpreter uses prelude-visible operations and ordinary library facilities.

The goal is not to eliminate Go or reimplement host arithmetic/I/O. It is to express Aiki evaluation semantics in Aiki.

---

## Cut III.0 — Runtime data model and environments

Define runtime structures for the interpreted program, likely using shapes such as:

```aiki
let @env [bindings, shapes, outer]
let @binding [name, value]
let @closure [params, rest_param, body, env]
```

Use a simple linear binding representation initially. A map/dictionary must not be added merely as a self-hosting convenience; if performance demonstrates a genuine language need, it should be proposed independently.

### Gate

Environment lookup, shadowing, update, enclosure, and closure capture behave according to documented Aiki semantics on focused tests.

---

## Cut III.1 — Evaluator: expressions and literals

Implement tree-walking evaluation for core expressions and values, including:

- numbers;
- strings/runes/symbols/booleans;
- lists;
- unary/binary operations;
- exact arithmetic through ordinary Aiki operations;
- indexing and field access;
- function literals as appropriate.

### Gate

Focused expression programs produce the same observable values/results under Go and Aiki evaluators.

---

## Cut III.2 — Evaluator: statements and control flow

Implement:

```text
let
assignment
if/else
while
return
match
block/program sequencing
```

### Gate

Focused control-flow programs agree behaviorally with the Go implementation and existing gold expectations.

---

## Cut III.3 — Functions, closures, recursion, and builtin bridge

Implement function application and lexical closure behavior.

Host/prelude operations are bridged by calling the actual visible Aiki/prelude binding rather than reimplementing their Go substrate behavior.

Conceptually:

```aiki
let call_builtin = (name, args) {
    let fn = env_get(prelude_env, name)
    apply(fn, args)
}
```

Tail-call optimization may initially be omitted unless conformance demonstrates that it is required for existing specified behavior.

### Gate

Function, closure, recursion, and prelude-use fixtures agree with the Go interpreter within documented stack limits.

---

## Cut III.4 — Modules

Initially prefer delegation of module loading to the host mechanism if that preserves the intended proof without introducing hidden evaluator semantics.

A later extension may self-host module loading using file read + lex + parse + eval.

### Gate

Programs importing representative Aiki modules execute equivalently under the self-hosted path, with the chosen delegation boundary explicitly documented.

---

## Cut III.5 — Behavioral conformance

Run the self-hosted interpreter against the executable behavior corpus.

Conformance now has three distinct surfaces:

```text
lexical conformance
syntactic conformance
behavioral conformance
```

Every failure is classified as:

- Aiki interpreter bug;
- Go implementation bug;
- specification ambiguity;
- unsupported but explicitly scoped behavior.

### Gate

The agreed acceptance corpus produces the same observable result/output/fault under both implementations.

---

## Cut III.6 — Self-interpretation

Feed the Aiki interpreter to itself:

```text
Go interpreter
  -> Aiki interpreter
      -> Aiki interpreter
          -> test program
```

### Gate

The nested interpretation produces the same specified observable result as direct execution for the selected proof program(s).

This is the final self-hosting proof.

---

# Cross-Phase Contracts and Invariants

## Grammar coupling

Where either Aiki self-host code or an editor declaration necessarily duplicates enumerated grammar facts, the repository should check them against the grammar authority. The Phase-I Cut I.0 token-authority extractor is the first concrete mechanism for this: it derives lexical facts from `grammar.ebnfx` and makes them available for comparison without generating or sharing lexer implementation code.

Examples:

```text
keyword inventory
operator inventory
newline policy
production coverage
GtkSourceView declarations
```

The independent algorithms remain independent; the authority does not drift.

## Diagnostic equivalence

As Phase II develops:

```text
CLI finding
  == language-service finding
  == LSP-projected finding
```

with equality defined semantically over source location, finding identity/message, and category/severity where applicable.

## Formatting equivalence

```text
aiki fmt source == language-service Format(source)
```

with the existing parse-preservation gate retained.

## Definition/symbol equivalence

```text
LSP definition location == tags definition location
```

because both consume the same symbol authority.

## Cross-implementation syntax equivalence

```text
reviewed token fixture
  == Go projection
  == Aiki projection

reviewed syntax fixture
  == Go projection
  == Aiki projection
```

Normalizers may remove representational noise; they must not hide semantic disagreement.

## Observation/instrumentation neutrality

Across all service capabilities:

```text
result(observer off, probe off)
  == result(observer on, probe off)
  == result(observer off, probe on)
```

except for explicitly requested observation/profiling output.

---

# Repository Shape

Exact names may evolve, but the dependency intent is approximately:

```text
engine/
  language/                 # service contract and implementation
    contract.go
    document.go
    diagnostic.go
    service.go
    symbols.go
    definition.go
    completion.go
    hover.go

  observe/                  # neutral observation/probe substrate
    ...

adapter/
  lsp/
  tags/

cmd/
  ...                       # aiki lsp, aiki tags, existing fmt/lint shells

extra/
  editors/
    xed/
      aiki.lang
      ...                   # only thin integration/config where needed
    nvi/
      README.md
    vscode/
      ...                   # thin LSP client/packaging

# location to be chosen during Phase I
selfhost/ or extra/selfhost/
  lex.ai
  parse.ai
  eval.ai
  ...

test/
  conformance/
    syntax/
      lex/
      newline/
      parse/
    behavior/
```

The precise location of self-hosted code should be decided from repository conventions during Phase I rather than guessed here.

---

# Acceptance Criteria

The combined proposal is complete when all three phases are gated.

## Self-description and conformance

1. An independent Aiki lexer and parser exist and run as ordinary Aiki programs.
2. They do not call the Go lexer/parser to perform their work.
3. Reusable reviewed lexical and syntactic conformance corpora exist.
4. The Go and Aiki implementations agree with those corpora across the accepted front-end surface.
5. Necessary duplicated grammar facts are executable-coupled to the grammar authority.

## Language services

6. An editor-independent language-service contract exists.
7. Language-service work is observable and instrumentable using neutral Aiki-owned contracts.
8. LSP is implemented as `aiki lsp`, with protocol concerns confined to an adapter.
9. Xed receives current lexical support and live semantic diagnostics without owning Aiki semantics.
10. nvi receives first-class Unix-native support through tags/lint/fmt rather than an artificial LSP layer.
11. VS Code is a thin client of the same LSP adapter.
12. CLI, LSP, tags, and editor declarations are executable-coupled to their common authorities where appropriate.

## Full self-hosting

13. An Aiki evaluator written in Aiki can execute the accepted Aiki behavior corpus.
14. It uses ordinary prelude/library capabilities and does not require new HAL primitives or Go-specific escape hatches.
15. No new value type is added solely to make the self-hosted interpreter convenient.
16. The Aiki interpreter can interpret itself interpreting at least one selected proof program with the specified result.
17. The Go implementation remains the bootstrap/reference/production runtime unless a future proposal deliberately changes that policy.

---

# Non-Goals

- Replacing Go as the bootstrap or production runtime.
- Eliminating the HAL or reimplementing genuine platform services in Aiki.
- Performance parity between the Go and Aiki interpreters.
- Turning nvi into an IDE.
- Making VS Code the architectural center of Aiki editor support.
- Making LSP synonymous with language services.
- Reusing the Go lexer/parser inside the Aiki front end merely to reduce code duplication; that would weaken the proof.
- Adding a map/dictionary, new value type, or builtin solely to make self-hosting easier.
- Requiring the initial Aiki parser to parse `grammar.ebnfx` dynamically. A grammar-driven Aiki parser may be a later, stronger experiment.
- Implementing every LSP feature. Capabilities are added only when justified by an Aiki service and a real client need.
- Embedding protocol-specific types in the service core.

---

# Open Questions

## Q1. What is the normalized syntax projection?

This is deliberately not fully specified here. Phase I should discover the smallest structure that both Go and Aiki parsers can emit naturally without privileging private implementation details.

The representation should be grammar-significant, stable enough for fixtures, and useful to later language-service/invariant consumers.

## Q2. How much grammar metadata should duplicated implementations consume directly?

The first Aiki parser should remain an independent hand-written implementation. However, keyword/operator/newline policy copies should be checked against grammar authority where practical.

A later grammar-driven Aiki parser would be a stronger but separate experiment.

## Q3. Where should the self-hosted implementation live?

Possible locations include `selfhost/`, `extra/selfhost/`, or another repository-consistent home. Decide during Phase I based on whether the implementation is treated as a conformance implementation, an example/application, or a durable language artifact.

## Q4. How should the language-service observer relate to `engine.Observer`?

The answer should come from actual dependency direction in Phase II. Extend an existing neutral contract only if that remains cohesive and does not force low-level packages upward. Otherwise use a sibling observer contract.

## Q5. What is the smallest practical Xed LSP integration?

Determine this against the actual Xed environment during Phase II. Prefer generic client/configuration support if available. If an Aiki-specific plugin is required, keep it thin and protocol-only.

## Q6. How should module loading work in the self-hosted evaluator?

Initial host delegation is acceptable if the boundary is explicit and the evaluator semantics remain independently implemented. Full self-hosted module loading can be a later extension.

## Q7. Is tail-call optimization required for initial conformance?

Do not implement it merely for completeness. Add it if the specified acceptance corpus or documented semantics require it.

---

## Q8. What unit is a source column?

Phase-I implementation exposed a real cross-implementation difference: the
current Go lexer advances columns by UTF-8 bytes, while ordinary Aiki string
indexing is rune-oriented. ASCII positions therefore agree, but non-ASCII
positions can diverge. Phase-I conformance fixtures are intentionally ASCII
until this policy is decided deliberately. Phase II retains the existing one-based UTF-8-byte position as the internal
compatibility surface and performs explicit protocol-boundary translation. The
LSP adapter advertises UTF-16 and converts byte columns to UTF-16 code units;
this avoids silently changing existing diagnostics while keeping editor positions
correct.

---

# Why This Fits Aiki

Aiki's existing architecture repeatedly favors explicit boundaries and executable agreement over hidden coupling:

- grammar over parser folklore;
- HAL over direct host dependence;
- observer over invisible behavior;
- probes over ad hoc measurement;
- executable documentation and gold files over untested claims;
- narrow tools and adapters over duplicated semantic authority.

The combined proposal extends the same pattern inward and outward.

**Inward**, Aiki expresses its own front end and eventually its evaluator, providing a genuinely independent implementation against which the Go implementation and specification can be tested.

**Outward**, Aiki exposes stable language knowledge through a contract that editors and Unix tools can consume without becoming new authorities.

The sequence matters:

```text
Aiki first demonstrates how its syntax can be independently apprehended.
Then it exposes that knowledge as a service.
Then it completes the stronger proof that Aiki can evaluate Aiki.
```

The result is not merely self-hosting and not merely editor support. It is a coherent architecture in which Aiki can **state, test, observe, and project its own language knowledge** without surrendering authority to any one implementation or client.
