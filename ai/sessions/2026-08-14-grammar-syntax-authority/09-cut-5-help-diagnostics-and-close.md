# Milestone 09 — Cut 5 help, diagnostics, and close

Status: GATED

Exposed newline policy through `help("newline")`, sourced from
`grammar.ebnfx` metadata, and pointed general help to that topic. The grammar
help wording was corrected to state the actual completion set and explicitly
note that `}` itself is not a terminating token.

The parser now preserves provenance for synthetic newline terminators. When
leftover input is a grammar-derived expression continuation after such a
terminator, the diagnostic explains that the previous newline ended the
statement and advises placing the continuation before the newline. No private
continuation list was introduced.

The three negative smoke diagnostics (`.`, `|>`, `+`) intentionally changed.
The final `enginesmoke` EBNF gold initially mismatched because grammar help
metadata is part of the structural dump. The user regenerated it with:

    ./aiki enginesmoke --stage ebnf --gold test/structure/engine

Inspection showed newline metadata/analysis and `GRAMMARHASH` movement while
production rule hashes remained unchanged. The final authoritative
`make validate` then passed:

- all Go tests passed;
- 408 Aiki tests passed;
- `smoke ok (46 tests)`;
- grammar coverage: 32 productions, 10 inputs;
- engine gold check: 10 inputs;
- treecheck: 471 files, 415 structurally justified, 56 explicitly allowed.

The proposal is implemented. Remaining newline-policy questions are deferred by
D2; recursive fmt/lint noise on negative fixtures is recorded as `buglist.md`
B1.
