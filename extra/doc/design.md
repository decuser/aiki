<!-- contract
allowed: NOW

Note: Every non obvious NOW claim should include a code anchor in brackets.
-->

# Aiki Design

## Layer authority

NOW syntax is structure only. [syntax package]
NOW semantics is meaning only and consumes syntax.Node. [semantics/eval/node.go EvalNode]
NOW runtime is capability only and hosts primitives. [runtime/hal package]
NOW tools are projections and contain no language rules. [tools package]
NOW resolver exists as a package and is not used for module loading. [resolver package]

## Grammar authority

NOW grammar source is embedded in syntax. [syntax/grammar_loader.go grammarSource]
NOW syntax.GetGrammar loads and caches the canonical grammar. [syntax/grammar_loader.go GetGrammar]
NOW syntax.Parse parses grammar source into a Grammar. [syntax package Parse]
NOW Grammar.ParseSource parses source into a syntax.Node. [syntax package Grammar.ParseSource]

## Evaluator coupling

NOW evaluator uses a handler map keyed by node.Type. [semantics/eval/node.go handlers]
NOW ValidateHandlers panics when a production lacks a handler. [semantics/eval/node.go ValidateHandlers]
NOW SetNodeGrammar stores grammar for import and load parsing. [semantics/eval/module.go SetNodeGrammar]

## Current module behavior

NOW import is an intrinsic implemented by evaluator. [semantics/eval/module.go evalImport]
NOW export is an intrinsic implemented by evaluator. [semantics/eval/module.go evalExport]
NOW import takes a string module name and one or more symbol names. [semantics/eval/module.go evalImport]
NOW module location is resolved by filesystem probes. [semantics/eval/module.go resolveModulePath]
NOW exported names are enforced only when export list is set. [semantics/eval/module.go evalImport env.GetExports]

## Operators and expressions

NOW infix operators are evaluated by evaluator logic, not environment lookup. [semantics/eval/node.go applyOp]
NOW operator shadowing is not implemented for infix syntax. [semantics/eval/node.go applyOp]

## Resources

NOW handle is a dedicated value type. [semantics/value/value.go type Handle]
NOW canvas is a dedicated value type. [semantics/value/canvas.go type Canvas]
NOW channel is a dedicated value type. [semantics/value/channel.go type Channel]

## Tools
NOW doclint subcommand checks documentation contract headers and tag usage for doc and work trees. [tools/doclint]
NOW smoke subcommand runs *_smoke.ai against .gold transcripts (IN/OUT/ERR/EXIT/DISPLAY) as canary contracts. [tools/smoke]

## IO behavior
NOW print writes arguments contiguously with no implicit separators; callers must print spaces and newlines explicitly. [runtime/hal print]

