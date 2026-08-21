# Proposal: Centralized Grammar Analysis

## Status

COMPLETE — implemented, integrated, and retained by the current full validation surface.

## Thesis

**Parse once. Derive structural facts once. Consume everywhere.**

`grammar.ebnfx` remains the sole authority for syntax. The grammar package should also be the sole authority for facts *derived from* that syntax. Consumers may interpret the grammar or compile derived facts into representations suited to their own work, but they must not independently walk the grammar tree to rediscover shared structural knowledge.

## Problem

The grammar-authority work eliminated several duplicated surface lists, but the current tree still contains multiple independent grammar walks:

- evaluator derives production `TokenRef` membership;
- evaluator separately derives terminal alternatives for `BINOP`;
- lint tests independently derive AST-producible node types;
- newline analysis is derived separately by parser and enginesmoke;
- formatter and other coverage checks read production structure independently.

This is not competing language authority—the grammar is still the source—but it is duplicated *interpretation machinery*. A grammar change can therefore make several consumers disagree about what the grammar says.

The redundancy is now concrete enough to remove.

## Governing boundary

Centralization applies to **derived structural facts**, not every use of the grammar tree.

The following remain consumer responsibilities:

- the parser's `parseExpr` interprets grammar expressions while parsing source;
- enginesmoke's expression rendering presents grammar expressions as text;
- the parser may compile newline declarations or analysis results into hash maps optimized for token filtering;
- evaluator policy remains responsible for parser-synthesized nodes such as `TERMINAL`;
- evaluator semantics remain responsible for what grammar-declared operators mean.

The rule is:

> Consumers may interpret or compile authoritative facts for their own purpose. They may not independently derive reusable structural facts already owned by grammar analysis.

## Design

### `grammar.Analysis`

Each parsed `Grammar` receives one cached analysis containing at least:

- production names;
- production-referenced token classes;
- AST-producible node types (productions + token references);
- terminal values reachable within each named production;
- cached newline analysis when the grammar has the required `expr`, `statement`, and newline declarations;
- the newline-analysis error, if derivation is unavailable.

Structural collection is performed centrally by one walk over the production expression trees. References are recorded, not recursively expanded, because every named production is itself visited exactly once.

FIRST/LAST/nullability/continuation derivation remains an internal grammar-analysis algorithm. "Analyze once" means one authoritative derivation per loaded grammar, not that recursive grammar algorithms must be forced into a single physical DFS.

### Lifetime

Normal grammar parsing builds the analysis after references are resolved. `grammar.Load` therefore returns an already-analyzed grammar.

The grammar remains mutable for tests and specialized tooling, so an explicit `Reanalyze()` operation is provided for deliberate post-load mutation. Ordinary consumers never call it.

### Newline analysis

`NewlineAnalysis` remains a distinct result because it depends on Aiki-specific production names and newline policy. It becomes a cached member of the general analysis rather than a computation independently requested by parser and enginesmoke.

If its prerequisites are absent or derivation fails:

- general structural analysis remains available;
- `Analysis.Newline` is nil;
- `Analysis.NewlineError` records the reason.

This keeps general grammar introspection usable without pretending every grammar must have Aiki's `expr` and `statement` productions.

### Consumer-specific compilation

Parser lookup structures such as `afterToken`, `afterLexeme`, `suppressDelta`, and continuation maps remain parser-owned. They are optimized representations of grammar-owned facts, not independent structural derivations.

## Cuts

### Cut 0 — proposal and inventory

Record the existing independent walkers and establish the boundary between structural derivation and legitimate interpretation/presentation.

### Cut 1 — central analysis model

Add `grammar.Analysis`, central structural collection, cached newline derivation, and tests for the current grammar. Attach the analysis to every parsed grammar. Preserve `AnalyzeNewlineRule()` temporarily as a compatibility accessor over the cache, not a recomputation.

### Cut 2 — structural consumers

Migrate evaluator coverage/operator membership, formatter coverage, and linter grammar-knowledge tests to consume `Grammar.Analysis`. Delete their local grammar walkers.

### Cut 3 — newline consumers

Migrate parser and enginesmoke to the cached `Analysis.Newline` / `NewlineError`. Parser may continue compiling lookup maps locally. Prove repeated access does not rederive grammar structure.

### Cut 4 — reconciliation

Search the tree for remaining independent structural grammar walkers. Keep only legitimate grammar interpretation/presentation paths. Update architectural decisions and session records. Run the complete validation gate.

## Acceptance criteria

1. A parsed/loaded grammar has one cached `Analysis`.
2. Structural collection of productions, `TokenRef`s, AST node types, and per-production terminals exists only in the grammar package.
3. Evaluator no longer defines `grammarTokenRefs` or `grammarTerminalAlternatives`.
4. Lint tests no longer define their own grammar AST walker.
5. Parser and enginesmoke do not independently invoke newline derivation.
6. `AnalyzeNewlineRule()`—if retained for compatibility—returns the cached result and never re-walks the grammar.
7. Parser grammar interpretation and enginesmoke expression rendering remain local and are explicitly recognized as non-duplicative responsibilities.
8. Parser-synthesized AST exceptions remain evaluator policy rather than being misrepresented as grammar facts.
9. Existing parsing, formatting, linting, evaluation, smoke behavior, grammar hashes/rule hashes, and Aiki-native tests remain behaviorally unchanged except for intentional structural-analysis representation changes, if any.
10. A final repository search finds no consumer-side grammar walkers deriving reusable structural facts.

## Non-goals

- A repository-wide application of this pattern to modules, builtins, documentation, or other authorities.
- Making the grammar immutable.
- Moving evaluator policy into the grammar package.
- Moving parser-specific lookup tables into grammar analysis.
- Eliminating grammar interpretation from the parser.
- Eliminating presentation traversal from enginesmoke.
- Redesigning newline syntax policy.

## Why this fits Aiki

The grammar-authority project established a single source for syntax. This proposal completes the next layer: shared knowledge derived from that source should itself have one derivation.

**Single authority for facts; local authority for meaning and policy.**
