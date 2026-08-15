# 01 — Grammar token authority extraction

Status: ACTIVE

## Intent

Implement Phase I, Cut I.0 from the active proposal: an ordinary Aiki program
that reads `engine/syntax/grammar.ebnfx`, locates the `@tokens` block, and
emits a deterministic normalized projection of the lexical authority needed by
the later independent lexer.

The extractor is intentionally not a full EBNF parser and does not generate the
lexer. Its role is executable coupling: duplicated lexical tables in the
independent implementation can later be checked against the grammar authority.

## Environment

The repository requires Go 1.24.0. This environment has Go 1.23.2 locally and
cannot download Go 1.24 because outbound network/DNS access is blocked.
Therefore the Aiki executable cannot be built here unless an existing compatible
binary becomes available. This limits executable validation but does not alter
the implementation contract.

## Implementation

Added:

- `selfhost/token_authority.ai` — Aiki-native narrow extractor for the
  `@tokens` block;
- `selfhost/token_authority.gold` — reviewed deterministic projection;
- `test/invariant/selfhost_token_authority_test.go` — executable invariant that
  builds Aiki, runs the extractor from the repository root, and compares its
  output with the reviewed projection.

The extractor emits four ordered facts: token classes, keywords, operators, and
delimiters. It derives all four from `grammar.ebnfx`; it does not generate or
share lexer algorithm code.

## Evidence

A disposable independent source-level extraction of the current `@tokens`
block produced the same four inventories as `token_authority.gold`.
`gofmt` completed on the added Go invariant test.

Attempted executable gate:

```text
GOTOOLCHAIN=local go test ./test/invariant -run TestSelfhostTokenAuthority -count=1
```

Result:

```text
go: go.mod requires go >= 1.24.0 (running go 1.23.2; GOTOOLCHAIN=local)
```

Gate status: ACTIVE — implementation complete; executable gate pending authoritative Go 1.24 toolchain.

## Decision

The durable independent implementation will live under top-level `selfhost/`.
It is neither an example nor an editor extra: it is a conformance
implementation of the language and will grow into the Phase-III interpreter.

### Follow-up: skip metadata

Cut I.3 established that parser normalization also needs grammar-owned `@skip`
metadata. The token-authority projection was therefore extended, rather than a
second list being introduced elsewhere, to emit `skips WHITESPACE COMMENT`.
The independent lexer and normalizer expose their copied skip inventory and the
Aiki coupling checks compare it to the grammar-derived projection.
