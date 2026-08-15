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
