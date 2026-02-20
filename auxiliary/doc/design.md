<!-- contract
allowed: NOW

Note: Every non obvious NOW claim should include a code anchor in brackets.
-->

# Aiki Design

## Layer authority

NOW syntax is structure only. [engine/syntax package]
NOW semantics is meaning only and consumes syntax.Node. [engine/semantics/evaluator/evaluator.go Eval]
NOW runtime is capability only and hosts primitives. [engine/runtime/hal package]
NOW tools are projections and contain no language rules. [interaction/tools package]
NOW resolver exists as a package and is not used for module loading. [resolver package]

## Grammar authority

NOW grammar source is embedded in definition package. [engine/syntax/definition/grammar.ebnf]
NOW definition.New returns GrammarContract as canonical loader. [engine/syntax/definition/definition.go New]
NOW GrammarContract provides GetTokens GetProduction GetStart. [engine/syntax/grammar_contract.go]
NOW definition/parser.go parses EBNF source into Grammar. [engine/syntax/definition/parser.go Parse]

## Evaluator coupling

NOW evaluator uses a handler map keyed by node.Type. [engine/semantics/evaluator/handlers.go handlers]
NOW ValidateHandlers panics when a production lacks a handler. [engine/semantics/evaluator/handlers.go ValidateHandlers]
NOW ev.SetGrammar stores grammar for import and load parsing. [engine/semantics/evaluator/evaluator.go SetGrammar]

## Current module behavior

NOW import is an intrinsic implemented by evaluator. [engine/semantics/evaluator/intrinsics.go evalImport]
NOW export is an intrinsic implemented by evaluator. [engine/semantics/evaluator/intrinsics.go evalExport]
NOW import takes a string module name and one or more symbol names. [engine/semantics/evaluator/intrinsics.go evalImport]
NOW module location is resolved by runtime filesystem probes. [engine/runtime/hal/substrate/go_runtime.go ResolvePath]
NOW exported names are enforced only when export list is set. [engine/semantics/evaluator/intrinsics.go evalImport scope.GetExports]

## Operators and expressions

NOW infix operators are evaluated by evaluator logic, not environment lookup. [engine/semantics/evaluator/operations.go applyOp]
NOW operator shadowing is not implemented for infix syntax. [engine/semantics/evaluator/operations.go applyOp]

## Resources

NOW handle is a dedicated value type. [engine/semantics/value/value.go Handle]
NOW canvas is a dedicated value type. [engine/semantics/value/value.go Canvas]
NOW channel is a dedicated value type. [engine/semantics/value/value.go Channel]

## Tools
NOW doclint subcommand checks documentation contract headers and tag usage for doc and work trees. [interaction/tools/doclint]
NOW smoke subcommand runs *_smoke.ai against .gold transcripts (IN/OUT/ERR/EXIT/DISPLAY) as canary contracts. [interaction/tools/smoke]

## IO behavior
NOW print writes arguments contiguously with no implicit separators; callers must print spaces and newlines explicitly. [engine/runtime/hal/substrate/go_runtime.go print]
