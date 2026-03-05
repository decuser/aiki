### Chat bootstrap for the lexer grammar refactor

Paste this at the top of a fresh chat.

Title
Aiki lexer refactor to grammar driven maximal munch

Goal
Replace the current lexer that hardcodes keyword and operator delimiter handling and discards whitespace with a new lexer that is fully driven by grammar token definitions, emits all tokens including whitespace newline comments, and supports maximal munch with tie break by grammar order.

Repo snapshot
I am working from this repo zip: `/mnt/data/aiki.zip`
Grammar lives in `engine/syntax` and is the single source of truth for token definitions.

Non negotiables

1. Lexer never discards anything. It only emits tokens.
2. Parser owns skipping, via tokens marked skip in grammar or an equivalent mechanism.
3. Keywords, operators, delimiters are tokens defined in grammar, not hardcoded lists.
4. Longest match wins. If tied, earlier grammar token wins.
5. Refactor is parallel, new lexer coexists with old until diff tests are green.
6. No behavior changes in eval. Only lex parse changes.

Work plan
Phase 0 add token dump observer and fixtures
Phase 1 implement new lexer behind a flag or alternate constructor
Phase 2 differential tests new vs old on smoke and gold corpus
Phase 3 integrate parser skipping so whitespace tokens can exist without breaking parse
Phase 4 flip default and keep old lexer as fallback for one tag

Acceptance tests

1. With filter mode that drops whitespace and comments from new stream, token sequence matches old lexer output for all existing tests.
2. With full mode after parser skipping, programs parse and eval the same as baseline.
3. Observer token dump is deterministic and stable.

Output artifacts

1. New lexer package or file set
2. Token dump observer implementation
3. Differential test harness and gold token dumps

Safety rails
No edits outside lexer parser grammar loader observer and tests. No import cleanup or unrelated formatting.

