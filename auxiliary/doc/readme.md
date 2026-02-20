<!-- contract
allowed: NOW PHIL
-->

# Aiki

PHIL Aiki is a minimal composable language.
PHIL Constraints force clarity.
PHIL Explicit over implicit.
PHIL Exactness by default.
PHIL Inspectability enables knowing.

## Use

NOW REPL: aiki
NOW Run file: aiki file.ai
NOW Eval expression: aiki -e 'print(1 + 2)'
NOW Format: aiki fmt file.ai
NOW Lint: aiki lint file.ai
NOW Smoke: aiki smoke

## Architecture

NOW syntax defines structure. [engine/syntax package]
NOW semantics defines meaning. [engine/semantics/evaluator package]
NOW runtime provides capabilities. [engine/runtime/hal package]
NOW tools project the language. [interaction/tools package]

NOW grammar is embedded and loaded by definition.New returning GrammarContract. [engine/syntax/definition/definition.go New]

NOW import currently resolves by filesystem lookup. [engine/semantics/evaluator/intrinsics.go evalImport]

## Docs

NOW design: doc/design.md
NOW decisions: doc/decisions.md
NOW style: doc/style.md
NOW todo: doc/todo.md
NOW done: doc/done.md
