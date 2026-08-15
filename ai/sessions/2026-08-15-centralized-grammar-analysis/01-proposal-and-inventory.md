# 01 — Proposal and inventory

**Status:** GATED

The current tree was inspected before modification. Independent reusable structural derivations were confirmed in evaluator (`grammarTokenRefs`, `grammarTerminalAlternatives`), lint tests (`grammarASTNodeTypes`), and repeated newline analysis calls from parser and enginesmoke.

The boundary was explicitly constrained: parser `parseExpr` is grammar interpretation and remains; enginesmoke `exprString` is presentation and remains; consumer-specific lookup maps remain consumer-owned.

Proposal added at `proposals/centralized-grammar-analysis.md`.
