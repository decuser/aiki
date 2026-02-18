<!-- contract
allowed: RULE NOW

Note: RULE statements reflect tool enforcement. NOW statements reflect runtime behavior.
-->

# Aiki Style

## Naming

RULE snake_case is the convention. [tools/lint]
RULE Mixed case should trigger lint diagnostics. [tools/lint]

RULE underscore prefix marks internal intent. [doc/design.md]

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
