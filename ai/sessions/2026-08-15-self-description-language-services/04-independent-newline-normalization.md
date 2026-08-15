# 04 — Independent newline normalization

Status: ACTIVE — implementation complete; authoritative gate pending.

## Result

Added `selfhost/normalize.ai`, an independent Aiki implementation of skip-token
filtering and grammar-declared physical-newline normalization. It preserves the
parser's delimiter-aware closer stack behavior, including mismatched/unmatched
closers, and emits synthetic `DELIMITER ;` tokens at the physical newline's
position when the grammar policy terminates a statement.

`selfhost/grammar_authority.ai` now also derives the narrow `@newline` authority:
newline token, completion token classes, completion lexemes, and suppression
pairs. `selfhost/check_normalization_authority.ai` executable-couples the Aiki
normalizer's copied tables to those grammar-derived facts and to `@skip`
metadata from the token block.

## Go seam

The existing parser-local normalization algorithm was extracted without semantic
change into `engine/syntax/normalize.go`. `syntax.NormalizeTokens` exposes the
normalized token surface; `NewParser` uses the same internal helper and retains
synthetic-terminator provenance for diagnostics. This is the smallest Go seam
needed to compare the two implementations rather than duplicating parser logic
inside a test.

## Conformance corpus

`test/conformance/syntax/newline/` covers ordinary statement termination,
parenthesis/bracket suppression, deliberately non-terminating grammar cases,
and mismatched closer behavior.

`test/invariant/selfhost_newline_conformance_test.go` requires the Go and Aiki
normalizers to match the same reviewed `.tokens` projections.

## Evidence

A disposable Go-1.23 syntax-only copy ran `go test ./engine/syntax/...` after the
normalization extraction; existing syntax and grammar tests passed. The
disposable non-graphics Aiki runtime then matched all four newline fixtures
exactly.

Authoritative Go-1.24 validation remains pending.
