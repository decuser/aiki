# Summary — Centralized Grammar Analysis

The grammar-authority project established `grammar.ebnfx` as Aiki's sole syntax authority. The post-merge audit showed a second-order duplication: several consumers still independently traversed grammar expressions to answer the same structural questions.

The resulting rule is narrower than "put everything in the grammar package":

> Parse once. Derive shared structural facts once. Consume everywhere.

The grammar package now owns one cached analysis with production names, TokenRefs, AST-producible node types, terminal values by production, and Aiki-specific newline analysis. Evaluator, formatter coverage, lint grammar-knowledge tests, parser diagnostics, and enginesmoke consume that view.

Consumer policy stays local. The parser still interprets EBNF expressions; enginesmoke still renders them; the parser still compiles lookup maps; evaluator still owns synthetic nodes and operator meaning. The centralization is therefore about factual derivation, not architectural symmetry.

This project also makes deliberate grammar mutation visible: loaded grammars are analyzed once, while tests that mutate a grammar must call `Reanalyze()` explicitly. That preserves the normal one-analysis lifetime instead of hiding recomputation inside consumers.
