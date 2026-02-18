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

## Architecture

NOW syntax defines structure. [syntax package]
NOW semantics defines meaning. [semantics/eval package]
NOW runtime provides capabilities. [runtime/hal package]
NOW tools project the language. [tools package]

NOW grammar is embedded and loaded by syntax.GetGrammar. [syntax/grammar_loader.go GetGrammar]

NOW import currently resolves by filesystem lookup. [semantics/eval/module.go resolveModulePath]

## Docs

NOW design: doc/design.md
NOW decisions: doc/decisions.md
NOW style: doc/style.md
NOW todo: doc/todo.md
NOW done: doc/done.md

## Tools
NOW `aiki doclint` validates documentation contracts in doc and work. [tools/doclint]

