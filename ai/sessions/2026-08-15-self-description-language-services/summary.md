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
