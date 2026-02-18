<!-- contract
allowed: HIST
-->

# Aiki Done

## v0.3.1 doclint
HIST Added `aiki doclint` subcommand enforcing documentation contracts for doc and work trees.
HIST Introduced doclint.ini to declare scope roots and valid tags.
HIST Added contract headers to design decisions manifesto style and work tracking docs.
HIST Split work docs into work/todo work/done work/backlog distinct from doc/.

## 2026 02 17 Grammar loader and package boundary stabilization

HIST Canonical grammar embedded via go embed. [syntax/grammar_loader.go]
HIST Introduced syntax.GetGrammar as sole public loader. [syntax/grammar_loader.go GetGrammar]
HIST Removed ambiguous Grammar constructor usage. [syntax package]
HIST Tools updated to use canonical loader. [tools/fmt tools/lint]
HIST Integration tests standardized to integration_test package. [semantics/integration]
HIST Resolved test and package import breakage after refactor. [go test ./...]
HIST Confirmed go test ./... passes cleanly. [makefile test]
HIST Grammar authority centralized in syntax. [syntax package]
HIST Tooling aligned to exported syntax API. [tools package]
HIST Circular import risk removed. [package structure]
HIST Structural boundaries clarified across syntax semantics runtime tools. [repository layout]

## v0.2.7 Architecture

### Grammar Cleanup
HIST Removed percent from OPERATOR and BINOP.
HIST Removed double equals from OPERATOR and BINOP.
HIST Removed not equals from OPERATOR and BINOP.
HIST Added modulo() to replace percent operator usage.
HIST Added equal() to replace double equals operator usage.
HIST Added not equal() to replace not equals operator usage.

### Grammar Evaluator Coupling
HIST Replaced evaluator switch statement with a handler map.
HIST Added ValidateHandlers which panics on missing handler.
HIST Called ValidateHandlers from SetNodeGrammar at startup.
HIST Eliminated drift by forcing evaluator updates when grammar changes.

### Eval Package Reorganization
HIST node.go contains handler map, EvalNode, and core eval logic.
HIST module.go contains SetNodeGrammar and import export.
HIST intrinsics.go contains call dispatch, apply, load.
HIST Removed dead code evalNodeImport and evalNodeExport.
HIST Deleted pragmatic directory.
HIST Removed LayerPragmatic constant.

## 2026 02 17 Grammar loader and package boundary stabilization

HIST Embedded canonical grammar in syntax.
HIST Canonical loader is syntax.GetGrammar.
HIST Tools updated to use GetGrammar.
HIST Integration tests standardized under semantics integration.
HIST Integration tests use integration_test package.
HIST go test ./... passes.

## v0.2.6 Parity

### EBNF Migration
HIST Lint rewritten with EBNF AST.
HIST Export statement verified.
HIST Import statement verified.
HIST Strict exports parses from strict.ai.

### Rich Errors Phase 1
HIST makeError(env, node, ...) includes file line source.
HIST Stack traces include from lines.
HIST HAL errors annotated with call site position.
HIST InspectAtLayer supports filtered display by layer.
HIST main frame pushed at entry points.
HIST Layer system implemented with user strict hal.
HIST Tests added in tests error_test.go.

## v0.2.5

### Architecture
HIST EBNF grammar driven parser with grammar.ebnf as source of truth.
HIST Lexer parser fmt lint derive from grammar.

### Internals
HIST Layer type added on StackFrame and Function.
HIST PushFrame and PopFrame manage call stack frames.
HIST AnnotateError decorates HAL errors.
HIST Lint rewritten with EBNF AST walker.
HIST strict.Exports parses from strict.ai with no hardcoded list.

## v0.2.4

### Syntax
HIST Bracket indexing works for list index, string index, and call results.
HIST Dot access is for fields only.
HIST Rest parameters supported with ...rest.
HIST Semicolon statement separator supported.
HIST GroupExpression preserves parentheses.

### Primitives
HIST ord(rune) converts rune to code point.
HIST apply(fn, list) spreads list as args.
HIST shell(cmd) runs shell command.
HIST Removed nth in favor of bracket indexing.

### CLI
HIST Added dash e flag for one liner expressions.
HIST quit() exits REPL.

### Internals
HIST HAL is single source of truth for builtins.
HIST REPL subcommand pattern implemented.
HIST Extracted version package.
HIST TrackingWriter replaced global state.

### Fixes
HIST Removed blank line on Ctrl C.
HIST Removed exit echo on Ctrl D.

## v0.2.3

HIST Lexer implemented as a state machine over UTF 8.
HIST Parser implemented as recursive descent with left to right evaluation.
HIST Evaluator implemented as tree walking.
HIST Eight types number boolean rune string bytes symbol list function.
HIST Rational arithmetic exact via big.Rat.
HIST Shapes support composition.
HIST Pattern matching implemented.
HIST Pipe operator supports error short circuit.
HIST Prelude includes map filter reduce range and hash map.
HIST File IO supports open create fread fwrite fclose.
HIST REPL uses readline.
HIST File runner implemented.
HIST Formatter preserves comments.
HIST Comma separators supported.
HIST Ruby style error messages.
HIST Canvas primitives via Ebiten.
HIST Concurrency spawn channel send recv.

