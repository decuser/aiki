<!-- contract
allowed: RULE NOW

Note: RULE statements reflect tool enforcement. NOW statements reflect runtime behavior.
-->

# Aiki Style

## Naming

RULE "_" is a valid identifier used for discard; it is exempt from case rules. [tools/lint]
RULE normal identifiers must use snake_case. [tools/lint]
RULE any mixedCase or PascalCase identifier must trigger a lint diagnostic. [tools/lint]
RULE a leading "_" marks internal intent only and has no semantic effect. [doc/design.md]


## Shadowing

RULE Shadowing prelude names should warn. [tools/lint]
RULE Shadowing intrinsics should error. [tools/lint and semantics/eval]

## Formatting

RULE Tabs for indentation. [tools/fmt]
RULE One statement per line after formatting. [tools/fmt]
RULE hash begins a comment. [tools/fmt and syntax lexer]

## Modules

NOW import resolves module strings through filesystem lookup. [semantics/eval/module.go resolveModulePath]
NOW export declares explicit exports for enforcement on import. [semantics/eval/module.go evalExport]

## Tooling
RULE Documentation files under doc and work must declare allowed tags in a contract header.
RULE `aiki doclint` must pass before changes are merged.

