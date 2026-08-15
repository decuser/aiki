# Phase I conceptual summary

The combined self-description/language-services effort began by making Aiki's
front end independently expressible in Aiki. The implementation confirmed the
proposal's central distinction: duplicate implementation is useful evidence,
while duplicated authority must remain executable-coupled to the grammar.

A narrow Aiki grammar reader derives lexical and newline policy facts from
`grammar.ebnfx`. The independent lexer and normalizer copy those facts where an
independent implementation naturally must, and Aiki programs compare the copies
against the derived authority. The lexer itself uses no regex or Go front-end
facility. The normalizer independently realizes skip/newline policy. The parser
is a hand-written recursive-descent implementation of the grammar.

Two existing Aiki semantics mattered materially. First, `equal()` is atomic and
intentionally non-structural for lists, so authority checkers require a small
recursive list comparison. Second, `and`/`or` are not short-circuit guards;
bounds-safe scanner code must not rely on C/Go-style boolean guard idioms.
These were discovered by running the new code rather than inferred away.

The syntax projection did not need a new format. Aiki already maintains a
human-readable, grammar-shaped parse surface in the engine `.parse.gold` files.
The second parser now targets those same reviewed artifacts. This gives a useful
triangulation: grammar/specification authority, Go front end, and Aiki front end
without inventing a second parse-gold authority.

A source-position issue surfaced for Phase II. The Go lexer advances columns by
UTF-8 bytes; Aiki string indexing is rune-oriented. ASCII conformance is exact,
but editor/LSP work must decide the language-level position unit deliberately
before non-ASCII positions are promised.

## VS Code client boundary

The final Phase-II editor confirms the adapter model rather than extending it.
VS Code registers `.ai`, supplies lexical presentation, and launches `aiki lsp`
through Microsoft's standard language client. `aiki.server.path` exists because
live Xed testing demonstrated that a desktop editor's environment must be
observed directly; an interactive shell PATH is not evidence of the editor's
PATH. Semantic behavior remains entirely in Aiki's language-service/LSP layers.


# Phase II completion and Phase III opening

Phase II closed on real-client evidence rather than protocol tests alone. Xed proved diagnostics and desktop-environment installation behavior; nvi proved a classic tags projection; VS Code exercised the wider LSP surface and exposed two useful protocol assumptions: nullable JSON-RPC results must be explicit, and editor caret positions cannot be assumed to coincide with identifier starts. The final VS Code development installer packages and installs a VSIX from a disposable out-of-tree staging area.

Phase III began by testing whether interpreted lexical state can itself be ordinary Aiki. A closure-backed environment proved sufficient: environment operations close over a mutable lexical binding variable while the binding collection remains persistent list data. This preserves shared closure capture and avoids introducing `store` or a dictionary merely for the interpreter.

The first evaluator work then produced a more fundamental result. Native Aiki source has an open symbol syntax, but the ordinary user/prelude surface has no dynamic string-to-symbol constructor. The substrate likewise has shaped-list construction (`_make_shaped_list`) only below the HAL boundary. A second interpreter cannot faithfully turn arbitrary symbol/shape lexemes into ordinary native values without either a general language-level construction facility or an invalid workaround. Self-hosting therefore did what it was intended to do: it found a sufficiency gap at the platform/language boundary before implementation accreted around it.

## Phase III module and conformance findings

The dynamic-value gap was closed with two deliberately small surface changes: `to_symbol("foo") -> :foo` adds the missing runtime constructor for symbols, while `shaped(:point, [1,2])` exposes an already-existing HAL shaped-list capability. The distinction matters: one is a genuine value-model completeness fix, the other removes an unnecessary access restriction.

Module loading then clarified the HAL boundary further. Delegating host `import()` was too weak because the resulting native module is opaque to runtime-name access in ordinary Aiki. The adopted design self-hosts every Aiki-source module end to end. A privileged bootstrap privately captures the HAL functions that blessed `lib/` source is entitled to use and configures interpreted module environments; it exports only a high-level `run(source, file_name)` capability, never raw `_` bindings. Thus source-module semantics remain independent while platform effects remain host capabilities.

Behavior conformance exposed a semantic distinction that focused probes had missed: `[@error, ...]` is an ordinary recoverable value, not an evaluator halt. The self-host evaluator now uses a private `:self_fault` control shape for its own halting propagation. Existing match behavior then agrees with the reference implementation, including matching recoverable error values. The behavior sweep also closed qualified module access in pipeline targets.

The representative behavior corpus now agrees across exact arithmetic, closures/recursion, pattern matching, pipelines, relative imports, pure and FFI-backed modules, Unicode regex positions, strings, bytes, hashes, newline policy, and file effects. Concurrency/select/spawn, debugger-only fixtures, interactive input/error handling, and graphics remain explicitly outside the current self-host proof rather than silently omitted.

The final self-interpretation attempt reached a performance boundary rather than a semantic contradiction. An outer self-hosted interpreter successfully entered the inner bootstrap and loaded the inner lexer and normalizer, then spent more than a minute in parser-self-parse of `selfhost/parser.ai` without producing a fault. Full nested interpretation therefore remains an open performance/proof-engineering problem: the next action is measurement and optimization, not host-parser substitution.
