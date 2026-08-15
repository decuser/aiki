# 02 — Central analysis model

**Status:** GATED

Added `engine/syntax/grammar/analysis.go`.

A parsed grammar now caches one structural `Analysis` after reference resolution. The central structural pass records production names, TokenRefs, AST-producible node types, and literal terminals per production. Newline analysis is derived once during the same analysis build and retained as `Newline` / `NewlineError`.

`Grammar.Reanalyze()` exists only for deliberate post-load mutation, primarily tests. Ordinary loaded grammars are analyzed once.

`AnalyzeNewlineRule()` was converted to a compatibility accessor over the cache.

Validation: grammar package tests pass in a disposable Go-1.23-compatible copy; cached pointer identity is tested.
