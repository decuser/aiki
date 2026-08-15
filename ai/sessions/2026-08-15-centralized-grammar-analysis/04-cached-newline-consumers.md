# 04 — Cached newline consumers

**Status:** GATED

Parser and enginesmoke no longer call newline derivation independently. Both consume the cached `Grammar.Analysis()` result.

Parser still compiles continuation membership into local maps for its diagnostic hot path. This is intentionally consumer-specific compilation, not structural rederivation.

`AnalyzeNewlineRule()` remains for compatibility/tests but returns the exact cached `NewlineAnalysis` pointer.

Validation: grammar and syntax package tests pass in the disposable local compatibility tree.
