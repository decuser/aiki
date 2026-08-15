# Proposal: The Grammar as Sole Syntax Authority Redux

## Status

Implemented on `grammar/newline-rule` in serial cuts and validated through the full repository gate on 2026-08-14.

## Implementation Outcome

The proposal was implemented without changing Aiki's parsing behavior through Cuts 1–4. Cut 5 intentionally changed only newline-related help and diagnostics. The final authoritative validation passed: all Go tests, 408 Aiki-native tests, 46 behavior smokes, grammar coverage across 32 productions and 10 inputs, engine structural golds, and treecheck.

Execution produced three material findings:

- the ambiguous followers `(`, `[`, and `-` already resolve in favor of a new statement after an inserted terminator; they do not fail;
- grammar analysis derives the overblocked leading continuations `* + . / < <= > >= and or |>`;
- `}` can end a complete expression but is not in the newline completion set, so a function literal can continue across a newline. A twelfth behavior smoke added in Cut 3 pins that existing behavior.

No syntax-policy change was made. Those findings are recorded in `docs/decisions.md` D2 for a later, separate language-design decision.

## Summary

Aiki is already unusually close to having a grammar-driven syntax implementation. The lexer is maximal-munch over `g.Tokens`, and the parser is a generic EBNF interpreter over `Sequence` / `Alternative` / `Repetition` / `Option` / `Terminal` / `Reference` / `TokenRef`. The productions themselves do not require a hand-written parser for individual Aiki constructs.

One important exception remains in the path from tokens to that generic parser, together with several unchecked boundaries around the grammar:

1. **Statement termination is a hidden surface rule.** `NewParser` runs an undeclared token-normalization pass that converts some newlines into synthetic `;` tokens, governed by a hardcoded `isComplete()` predicate. This rule is user-visible and appears nowhere in the grammar.

2. **Grammar/implementation coupling is incompletely checked.** `validateHandlers` establishes part of the grammar/evaluator relationship, but the reverse relation is not established. It also iterates lexical token definitions more broadly than the set of AST nodes the grammar can actually produce. No production references `KEYWORD`, `OPERATOR`, `DELIMITER`, or `NEWLINE`, yet those categories participate in the current accounting.

3. **The formatter and linter contain unchecked structural knowledge.** The formatter has six production cases intentionally handled by their parents — `field`, `select_case`, `select_default`, `param_list`, `rest_param`, and `literal` — but nothing asserts that relationship. The linter independently recognizes grammar structure without a corresponding coupling check.

4. **Surface membership is duplicated outside the grammar.** Most visibly, `BINOP` membership is declared in `grammar.ebnfx` and restated in evaluator code.

The objective is not to change how Aiki parses. It is to make `grammar.ebnfx` the complete and authoritative declaration of Aiki's surface rules, and to make implementation knowledge that overlaps that declaration mechanically checkable rather than capable of silent drift.

## Governing Idea

Two claims must remain deliberately separate:

- **The grammar is the sole authority on surface.** What forms exist, which tokens may appear where, and the declared rule governing where a newline terminates a statement.
- **The evaluator is the sole authority on meaning.** That `+` adds is semantics and belongs in evaluator code. It cannot and should not move into the grammar.

The system should enforce the **boundary between them**. Whenever implementation code depends on a set or structural fact already declared by the grammar, either that fact should be obtained from the grammar or the coupling should be checked mechanically.

This does not mean every implementation switch must disappear. It means no second, unchecked statement of Aiki's surface should quietly become authoritative.

## The Defect, Demonstrated

The grammar declares:

```text
pipe_expr = infix_expr { "|>" postfix_expr }
```

Taken alone, this production permits a pipeline to continue across a newline in either direction.

The token-normalization pass does not. `engine/syntax/parser.go` currently contains:

```go
// Go-style rule: insert ";" after complete token when followed by newline,
// but not inside () or [] ...
if tok.Type == "NEWLINE" {
    if parenDepth == 0 && len(filtered) > 0 && isComplete(filtered[len(filtered)-1]) {
        // insert synthetic ;
    }
}
```

and:

```go
func isComplete(tok Token) bool {
	switch tok.Type {
	case "NAME", "NUMBER", "STRING", "RUNE", "SYMBOL":
		return true
	}
	switch tok.Lexeme {
	case ")", "]", "true", "false":
		return true
	}
	return false
}
```

Outside suppressed delimiters, a newline following one of these tokens becomes `;`. Since `NAME` completes a statement, a line ending in a bare name is terminated before a following `|>` reaches the grammar parser.

### Reproduction

Confirmed on Linux, `v0.4.0-alpha`.

`break-before.ai` — operator trailing. **Passes**, prints `[2, 4, 6]`:

```text
use("list")

let data = [1, 2, 3]
let double = (n) { return n * 2 }

let out = data |>
	map(double)

println(out)
```

`break-after.ai` — operator leading. **Fails**:

```text
use("list")

let data = [1, 2, 3]
let double = (n) { return n * 2 }

let out = data
	|> map(double)

println(out)
```

```text
parser: /tmp/break-after.ai:7:1:
|> map(double)
^
expected end of input, got '|>'
```

The failure is consistent with the normalization mechanism: by the time the generic parser sees the stream, the newline after `data` has already become a hard boundary.

Aiki already has `newline_smoke.ai`, which exercises permitted multiline continuations such as trailing operators, open delimiters, and trailing pipes. It does not pin the rejected leading-continuation cases or several interactions of the normalization rule. Those become explicit Cut 0 probes.

## Cut 0 Probe Set

The proposal requires **eleven probes**.

The first eight establish the main newline boundary behavior.

### 1. `break-before.ai`

Trailing pipeline operator. Expected to **pass**.

```text
use("list")

let data = [1, 2, 3]
let double = (n) { return n * 2 }

let out = data |>
	map(double)

println(out)
```

### 2. `break-after.ai`

Leading pipeline operator. Expected to **fail**.

```text
use("list")

let data = [1, 2, 3]
let double = (n) { return n * 2 }

let out = data
	|> map(double)

println(out)
```

### 3. `break-access.ai`

Leading access operator. Expected to **fail**.

```text
use("string")

let s = string
	.upper("abc")

println(s)
```

### 4. `break-binop.ai`

Leading binary operator. Expected to **fail**.

```text
let x = 1
	+ 2

println(x)
```

### 5. `break-paren.ai`

Newline inside parentheses. Expected to **pass** because delimiter suppression applies.

```text
use("list")

let data = [1, 2, 3]

println(length(
	data
))
```

### 6. `break-index.ai`

Leading index. **Observed to pass**: newline insertion terminates `let y = x`, and `[1]` is parsed as a new expression statement. `println(y)` therefore prints the original list. This is the current ambiguity resolution and must remain unchanged through Cuts 1–4.

```text
let x = [1, 2, 3]
let y = x
	[1]

println(y)
```

### 7. `break-call.ai`

Leading call. **Observed to pass**: newline insertion terminates `foo()`, and `(bar)` is parsed as a new grouped-expression statement. This is the current ambiguity resolution and must remain unchanged through Cuts 1–4.

```text
let foo = () { return (n) { return n } }
let bar = 1

foo()
(bar)
```

### 8. `break-unary.ai`

Leading unary `-`. **Observed to pass**: newline insertion terminates `x`, and `-1` is parsed as a new expression statement. This is the current ambiguity resolution and must remain unchanged through Cuts 1–4.

```text
let x = 5
x
-1
```

Three additional probes pin interactions that Cut 2 could otherwise change inadvertently.

### 9. `newline-blank.ai`

Multiple blank physical lines between statements.

`NEWLINE` is currently tokenized as `/\n+/`, so a run of blank lines becomes one newline token and results in one termination event. The precise observable behavior must be captured before the filter loop changes.

Example:

```text
let x = 1


let y = 2

println(x + y)
```

Expected to **pass**.

### 10. `newline-comment.ai`

Trailing comment before newline.

`COMMENT` is `@skip`, so the comment is removed before newline normalization examines the preceding meaningful token. This behavior should be pinned because Cut 2 rewrites exactly that path.

```text
let x = 1 # comment
let y = 2

println(x + y)
```

Expected to **pass**.

### 11. `newline-block.ai`

Newline handling inside braces.

The proposed `@help` text explicitly states that braces do not suppress newline termination, so that assertion needs executable evidence.

```text
use("list")

let double = (n) { return n * 2 }

let f = () {
	let data = [1, 2, 3]
	let out = data |>
		map(double)
	return out
}

println(f())
```

Expected to **pass**.

All eleven probes must be added to `test/behavior/` with golds based on **observed baseline behavior**. Failure probes must pin the relevant diagnostic text and exit behavior.

Their first purpose is observational. Only after that baseline exists should any result be characterized as intentionally preserved.

## What the Inserted `;` Is Actually For

The grammar uses `;` optionally in two places:

```text
program = { statement [ ";" ] }
block   = "{" { statement [ ";" ] } "}"
```

So `;` is not generally required merely to distinguish adjacent statements. For example, `let x = 1` followed by `let y = 2` can terminate naturally when the expression parser has no valid continuation and the enclosing repetition proceeds to the next statement.

The inserted `;` serves a narrower and more important purpose: it can **force termination where the grammar would otherwise continue a greedy expression parse**.

Two obvious continuation sites are:

```text
postfix_expr = primary { call | index | access }
pipe_expr    = infix_expr { "|>" postfix_expr }
```

and infix expression structure introduces additional continuation tokens.

Without a hard boundary, forms such as:

```text
foo()
(bar)
```

can potentially be interpreted as one postfix expression rather than two statements.

The grammar therefore currently understates the role of `;`. It presents an optional syntactic element while token normalization uses a synthetic `;` as a statement-termination barrier.

That relationship needs to be declared explicitly.

## Continuation and Ambiguity

The current grammar yields two useful conceptual sets:

- **FIRST(statement)** — tokens that may begin a statement.
- **CONTINUATION(expr)** — tokens that may legally continue an already complete expression.

The candidate continuation set includes:

```text
|>  .  (  [  +  -  *  /  <  >  <=  >=  and  or
```

The apparent intersection with `FIRST(statement)` is:

```text
(  [  -
```

This distinction matters:

- `(`, `[`, and `-` can both begin a new statement and continue an existing expression. A newline before them is structurally ambiguous.
- `|>`, `.`, `+`, `*`, `/`, `<`, `and`, and similar tokens cannot begin a statement. If the grammar permits them as continuations, terminating immediately before them rejects a form with only one grammatical interpretation.

The present normalizer does not inspect the following token. It asks only whether the token *before* the newline looks complete. Consequently it handles the two classes differently. For ambiguous continuations such as `(`, `[`, and `-`, insertion of `;` chooses the new-statement interpretation. For unambiguous continuation tokens such as leading `|>`, `.`, and `+`, the same insertion rejects the only grammatical continuation.

This proposal **does not change that behavior**.

The current convention may be a desirable Go-shaped style rule. The immediate defect is that the convention is hidden.

Whether Aiki should later permit unambiguous leading continuations is a language-design question and belongs in a separate proposal.

## Goals

1. `grammar.ebnfx` states Aiki's complete surface contract, including its newline-termination policy.
2. `engine/syntax/parser.go` contains no hardcoded Aiki lexeme or token-type membership lists governing newline termination.
3. A missing or malformed newline declaration is detected when the grammar is loaded.
4. The declared newline rule is mechanically compared with the productions rather than merely trusted.
5. Grammar/implementation couplings are enforced in both directions where bidirectional coverage is meaningful.
6. Surface membership already represented by a grammar category is not independently restated in Go without an explicit checked reason.
7. The existing 32 productions remain readable. Newline mechanics do not spread through expression productions.
8. Observable parsing behavior remains unchanged throughout Cuts 1–4.

## Non-goals

- Changing which existing Aiki programs parse.
- Deciding whether leading `|>`, `.`, or binary operators should become legal.
- Moving semantics into the grammar. `operators.go` continues to own the meaning of operators.
- Introducing `NEWLINE` throughout the productions.
- Making the EBNF format a general-purpose grammar system for unrelated languages.
- Redesigning the REPL.
- Redesigning diagnostics beyond the parse-leftover path directly exposed by this work.
- Requiring every subsystem to contain a dedicated handler for every grammar production when generic traversal is itself intentional behavior.

## Design

### 1. Declare the Newline Rule in `grammar.ebnfx`

Add one block adjacent to `@tokens`, outside the productions:

```text
@newline {
    token         NEWLINE
    after_token   NAME NUMBER STRING RUNE SYMBOL
    after_lexeme  ) ] true false
    suppress_in   ( )
    suppress_in   [ ]

    @help "A newline ends the current statement when the preceding token can
           complete one, except inside ( ) or [ ]. Braces do not suppress:
           a block contains statements, so newlines inside it terminate them."
}
```

The exact directive syntax may be adjusted during implementation if required by the grammar-file parser, but the information represented by it should remain explicit.

The `token` directive names the lexical token that represents a physical newline. This is required if `parser.go` is to contain no Aiki-specific token name: `@newline` must declare not only when a newline terminates a statement, but which token is the newline.

There are intentionally two completion lists.

`isComplete` currently distinguishes:

- token *types*: `NAME`, `NUMBER`, `STRING`, `RUNE`, `SYMBOL`;
- literal *lexemes*: `)`, `]`, `true`, `false`.

Those are different categories of information. Collapsing them into one apparent token set would obscure how the lexer represents the values involved.

`SHAPE` is absent from the current completion set. Whether that absence is guaranteed by the grammar or merely harmless under the present productions is precisely the kind of relationship the validation step should expose rather than assume.

### 2. Represent and Parse the Declaration

A likely representation in `engine/syntax/grammar/grammar.go` is:

```go
// NewlineRule declares how physical newlines become statement boundaries.
type NewlineRule struct {
	Token       string
	AfterToken  []string
	AfterLexeme []string
	SuppressIn  [][2]string
	Meta        Meta
	Pos         engine.Position
}

type Grammar struct {
	Tokens      []TokenDef
	Productions map[string]*Production
	Start       string
	Newline     *NewlineRule
}
```

The exact existing `Grammar` shape should be respected during implementation; this is architectural intent rather than a requirement to force an incompatible representation.

`engine/syntax/grammar/ebnf.go` should parse `@newline` alongside the other grammar-level declarations, using the existing directive/decorator parsing machinery where appropriate.

A built-in behavioral default is specifically undesirable. If the grammar requires newline normalization and the declaration is absent, grammar loading should fail clearly. Otherwise the hidden rule merely moves from `parser.go` into grammar-loading code.

The loader should also validate the declaration itself:

- named token types exist;
- named literal lexemes are recognized by the grammar token vocabulary;
- suppression pairs are structurally valid;
- duplicate or contradictory entries are diagnosed.

### 3. Make the Parser Consume the Declaration

The free function:

```go
isComplete(...)
```

must disappear.

The normalization pass should instead consume a compiled representation of `g.Newline`. Token-type and lexeme membership can be materialized as maps when the parser is constructed rather than scanned repeatedly.

Suppression depth should likewise be driven by the declared pairs rather than hardcoded cases for `(` and `[`.

The important acceptance test for this cut is structural:

> there is no second newline-completion vocabulary in `parser.go`.

The parser may contain generic machinery for applying a newline rule. It must not contain Aiki's particular newline rule.

### 4. Validate the Newline Rule Against the Productions

This requires more than a simple FIRST-set intersection because the implemented rule operates on the token before the newline while ambiguity is exposed by the token after it.

The validator should derive and report at least three structural facts.

#### 4a. Possible Expression Endings

Derive the token classes or literal terminals that can occur at the end of a complete expression in statement context.

Conceptually:

```text
END(expr)
```

This lets the system ask whether the declared `after_token` / `after_lexeme` policy corresponds to the expression endings for which newline termination is relevant.

#### 4b. Possible Expression Continuations

Derive tokens that may continue an expression after a point at which the preceding portion is already complete:

```text
CONTINUATION(expr)
```

This includes continuation introduced through postfix, pipeline, infix, and any other reachable repetition or option that extends an expression.

Traversal must respect statement-context boundaries such as blocks rather than treating every nested repetition as continuation of the outer expression.

#### 4c. Ambiguous Continuations

Derive:

```text
AMBIGUOUS = FIRST(statement) ∩ CONTINUATION(expr)
```

These are tokens for which a newline can separate two statements or continue one expression.

The validator can then characterize the declared rule in terms of the grammar rather than pretending the next-token set alone proves soundness.

Useful findings include:

- **uncovered endings** — grammar-derived complete-expression endings for which the declared rule does not impose the intended termination behavior;
- **ambiguous continuations** — continuation tokens that can also begin statements;
- **overblocked continuations** — continuation tokens that cannot begin a statement but are nevertheless rejected whenever the current preceding-token rule inserts a terminator;
- **invalid declaration entries** — newline metadata that names things unsupported by the grammar.

One invariant matters:

> A startup panic must be based on a clearly stated structural invariant, not on a heuristic approximation mislabeled as proof.

For that reason Cut 3 has an explicit analysis gate: compute the sets against the baseline grammar first. If the current rule does not satisfy the proposed hard invariant, stop the cut and record the discovery. Do not silently change Aiki's behavior merely to make the validator pass.

Once an invariant is confirmed, violation of it may become a grammar-load or startup failure.

`overblocked continuations`, by contrast, are not failures. They are evidence describing the cost of the current language rule.

### 5. Tighten Grammar/Evaluator Coverage

#### 5a. Handler Coverage in Both Directions

The existing `validateHandlers` relationship should be expressed in terms of AST nodes the parser can actually produce.

Grammar → evaluator remains essential:

> every reachable AST node requiring evaluator dispatch has an evaluator treatment.

The reverse direction should also be checked:

> every evaluator dispatch key corresponds to a reachable grammar-produced node or to an explicitly documented synthetic node.

Synthetic exceptions should be explicit rather than blurred into lexical token coverage.

#### 5b. Distinguish Token Definitions from AST-Producing Token References

A token being defined in `@tokens` does not imply that the parser emits it as an AST node.

This is already confirmed for:

```text
KEYWORD
OPERATOR
DELIMITER
NEWLINE
```

No production references any of those token definitions.

Coverage should therefore be based on grammar reachability and actual parser representation, not merely the lexical token-definition set.

Likewise, use of `NEWLINE` by the `@newline` declaration does not make it an AST-producing `TokenRef`.

**Lexical-policy reachability and AST reachability are separate relations.**

That distinction should be encoded.

### 6. Formatter Coverage

`printNode` currently has a generic fallback. Generic recursion is sometimes correct, but for formatter nodes it can also erase syntax when a leaf or token value has no child through which to recur.

The formatter therefore needs an introspectable coverage model.

Six productions are currently intentionally handled by their parents:

```text
field
select_case
select_default
param_list
rest_param
literal
```

That relationship is already established from the current formatter and should become explicit rather than being rediscovered during execution.

A map of production/node names to formatting functions is one possible implementation. The required invariant is:

- every reachable node whose formatting requires explicit syntax emission has a formatting treatment;
- parent-handled nodes are listed explicitly;
- generic fallback cannot silently discard a leaf token or production whose textual value matters;
- every explicit formatter dispatch key corresponds to a reachable node or documented synthetic case.

The formatter bug history makes this a high-value coupling to enforce.

### 7. Linter Coverage

The linter is different from the formatter and should not be forced into the same model.

A linter may legitimately traverse most AST nodes generically and attach special behavior only to a subset. Requiring a dedicated rule handler for all productions would manufacture boilerplate rather than enforce correctness.

Instead check two things:

1. every node name explicitly recognized by the linter corresponds to a reachable grammar node or documented synthetic node;
2. generic traversal is structurally complete, so introducing a new production cannot create an unvisited subtree merely because no linter rule names it.

Where the linter maintains grammar-sensitive sets — especially scope-forming constructs — those sets should either be derived from an authoritative declaration or covered by explicit invariant tests.

The lesson of the previous block-scope drift is not that the linter lacked handlers for unrelated productions. The defect was that it independently asserted a structural fact that no coupling checked.

### 8. Remove Duplicated `BINOP` Membership

The grammar declares:

```text
BINOP = "+" | "-" | "*" | "/" | "<" | ">" | "<=" | ">=" | "and" | "or"
```

Evaluator helper code currently restates that membership.

These are different questions:

- **Is this token a binary operator?** — syntax; the grammar already answers it.
- **What does this operator mean?** — semantics; evaluator code answers it.

The membership test should therefore be obtained from the grammar's `BINOP` alternatives or grammar-derived data compiled at startup.

`operators.go` may continue to dispatch on concrete operators. Add a bidirectional invariant:

- every `BINOP` grammar alternative has semantic treatment;
- every binary-operator semantic treatment corresponds to a `BINOP` alternative.

Audit `"true"` / `"false"` and unary `"not"` / `"-"` for the same *kind* of duplicated membership, but change them only where the grammar provides a clean authoritative category and the resulting coupling is clearer than the code it replaces.

The principle is elimination of competing surface authorities, not elimination of all string literals from semantic code.

### 9. Diagnostics

The current leftover-token error is less informative than production-backed parser failures because the grammar parser has already completed one valid parse before discovering unconsumed input.

Once newline termination is declared, the normalization pass can retain enough provenance to distinguish a real or synthetic statement boundary.

When the remaining token is a known continuation rejected by newline termination, the diagnostic should explain the surface rule rather than merely report unexplained trailing input.

For example:

```text
parser: /tmp/break-after.ai:7:1:
|> map(double)
^
the previous newline ended the statement.
'|>' cannot begin a statement under Aiki's current newline rule.

write the continuation before the newline:

    let out = data |>
        map(double)
```

The wording should be generated from or coupled to the `@newline` declaration rather than introducing another independent list of offending lexemes.

The exact final text belongs to Cut 5 and will intentionally alter the diagnostic gold for affected failures. That is a documented behavior change in diagnostics only, not parsing.

## Phased Implementation

Each cut is independently recorded and validated. Cut 0 occurs first without exception.

### Cut 0 — Baseline and Probes

Capture the baseline on the new authoritative tree.

Run and record the applicable existing gates:

```text
go test ./...
aiki test ./...
behavior smokes
engine gold checks
grammar coverage
race checks where applicable
treecheck
```

Record exact commands, environment limitations, and results in the session milestone.

Add the eleven newline probes to `test/behavior/` with golds based on **observed** current behavior.

Run the relevant gates again.

Nothing about the language implementation changes.

`GATED` when:

- the pre-probe baseline is recorded;
- all eleven probes reflect actual baseline behavior;
- adding those tests introduces no unexpected regression;
- the dated session record identifies the exact restart point.

### Cut 1 — Declare, Unused

Add:

- `NewlineRule`;
- `@newline` parsing;
- declaration validation;
- the `@newline` block in `grammar.ebnfx`.

Parse and retain the rule but do not use it for token normalization.

Update only expected structural/gold artifacts caused by the grammar metadata.

Observable Aiki parsing behavior must remain unchanged.

### Cut 2 — Consume, Delete

Make newline normalization consume `g.Newline`.

Delete `isComplete` and the hardcoded suppression/completion vocabulary from `parser.go`.

This is the load-bearing behavioral-equivalence cut.

All eleven newline probes and all existing behavior golds must remain unchanged from the Cut 0 observational baseline.

Any parse-result movement is a transcription error or a newly discovered pre-existing inconsistency. Stop and investigate rather than reblessing it.

### Cut 3 — Derive and Validate

Implement grammar analysis for:

- expression endings;
- expression continuations;
- ambiguous continuations;
- overblocked continuations;
- declaration validity.

Run it against the actual grammar before enabling any new hard failure.

Record the derived sets in the milestone and in `docs/decisions.md`.

If the current newline rule violates the intended invariant, **stop the cut** and record the finding. Decide separately whether the invariant is wrong or Aiki's behavior needs a language change.

Only after the invariant is shown to describe current intended behavior should violation become a startup or load failure.

### Cut 4a — Evaluator Coupling

Refine `validateHandlers` around reachable AST nodes.

Add reverse coverage and explicit synthetic-node exceptions.

Remove handlers shown to be unreachable only after confirming they cannot be emitted by the parser.

Gate independently.

### Cut 4b — Formatter Coupling

Make formatter coverage introspectable.

Encode the six confirmed parent-handled productions explicitly.

Prove that leaf/token syntax cannot disappear silently through fallback traversal.

Gate independently.

### Cut 4c — Linter Coupling

Check linter dispatch names against reachable AST nodes and verify generic traversal completeness.

Identify and gate any grammar-sensitive structural sets maintained by the linter.

Do not require meaningless one-handler-per-production symmetry.

Gate independently.

### Cut 4d — Operator Membership

Make `BINOP` membership grammar-derived.

Add bidirectional `BINOP` ↔ semantic-operator coverage.

Audit boolean/unary membership duplication and change only where the grammar supplies a clean authority.

Gate independently.

### Cut 5 — Help and Diagnostics

Expose the newline policy through `aiki help`, sourced from `@newline` metadata.

Improve the leftover-token diagnostic where newline normalization caused the boundary.

This cut may intentionally change **diagnostic text only**. Rebless affected failure golds deliberately and record the exact change.

## Acceptance Criteria

1. `grammar.ebnfx` contains an explicit newline policy.
2. A grammar requiring that policy but lacking a valid `@newline` declaration fails clearly at load.
3. `engine/syntax/parser.go` contains no hardcoded Aiki token-type or lexeme membership list governing newline completion or suppression.
4. `isComplete` no longer exists.
5. All eleven newline probes preserve their Cut 0 parse success/failure behavior through Cuts 1–4.
6. The grammar can derive and report expression endings, continuation tokens, and ambiguous continuations sufficiently to check the declared newline policy.
7. Any hard newline-policy invariant enforced at startup has first been demonstrated against the baseline grammar.
8. `validateHandlers` is bidirectional over the actual AST-producing node set, with explicit exceptions for synthetic nodes.
9. `KEYWORD`, `OPERATOR`, `DELIMITER`, and `NEWLINE` remain recognized as lexical definitions rather than AST-producing grammar references unless the grammar itself changes.
10. Lexical token definitions, newline-policy references, and AST-producing token references are not conflated.
11. Formatter coverage is mechanically checked, including the six explicit parent-handled cases and protection against silent leaf loss.
12. Linter-specific node names are checked against reachable AST structure and generic traversal completeness is tested.
13. `helpers.go` no longer independently restates `BINOP` membership.
14. `BINOP` alternatives and binary semantic implementations are checked in both directions.
15. `help("newline")` exposes the newline policy from grammar metadata, and general help points to that topic.
16. Except for the intentional Cut 5 diagnostic improvement, existing observable behavior remains unchanged.
17. No expected test, behavior-gold, structural-gold, or treecheck difference remains unexplained.

### Acceptance Reconciliation

All acceptance criteria are satisfied. Criterion 5 refers to the eleven Cut 0 baseline probes; Cut 3 added a twelfth smoke, `newline_function_end_smoke.ai`, after grammar analysis exposed the uncovered `}` expression ending. That additional smoke preserves and documents existing behavior rather than changing the Cut 0 baseline contract.

No hard newline-policy soundness invariant was added under criterion 7 because the derived grammar facts demonstrate that the present rule both overblocks some unambiguous continuations and leaves `}` uncovered. D2 records the explicit decision to preserve those language-policy choices in this authority refactor rather than change syntax to satisfy a validator.

## Validation Couplings

Affected couplings include:

- `grammar.ebnfx.ebnf.gold` or equivalent grammar-structure gold;
- engine golds that serialize grammar structure;
- grammar coverage;
- behavior smokes;
- executable documentation;
- `aiki test ./...`;
- evaluator coverage/invariant tests;
- formatter invariant tests;
- linter traversal/coverage tests;
- `treecheck` and `treecheck.allow` where new files require declaration.

### Behavioral Equivalence

Do **not** require raw command output to be byte-identical. Full test output can contain timing, temporary paths, test counts, ordering, and other incidental differences.

The equivalence criterion is instead:

- same parse success/failure result;
- same program output;
- same behavior-gold contents;
- same applicable structural-gold contents except where the proposal intentionally adds grammar metadata;
- no new unexpected test failures;
- no unexplained validation delta.

Where byte identity is meaningful — for example a specific `.gold` file or captured program stdout — compare that artifact directly.

## Documentation

`docs/decisions.md` is Aiki's durable architectural decision record and already contains D1.

At Cut 3, add a decision entry recording:

- the declared newline rule;
- the derived expression-ending set;
- the derived continuation set;
- the ambiguous continuation set;
- the overblocked continuation set;
- whether any hard startup invariant was adopted and why;
- the explicit decision not to change leading-continuation syntax in this proposal.

The proposal remains the design intent. The `ai/` session record preserves discoveries, deviations, validation, and the reason the implemented design ended where it did.

## Risks

### Validator Overclaim

The largest conceptual risk is calling an approximation a proof.

FIRST sets alone are insufficient because newline normalization acts on the preceding token while ambiguity concerns possible following continuations. The validator must model enough of both relationships to support whatever invariant it claims.

If that cannot be done cleanly, report useful derived facts rather than installing a false hard guarantee.

### Backtracking Cost

No production changes are planned, so the parser's grammar search space should remain unchanged.

Newline normalization remains linear over the token stream. Compile declaration membership into efficient lookup structures if useful.

Measure with the existing profiler after implementation rather than relying solely on expectation.

### Hidden Current Behavior

The additional probes may reveal behavior different from the expectations in this proposal.

That is a successful Cut 0 discovery, not a failure of the work. Record the observed behavior and reconcile the execution plan before Cut 1.

### Error-message Regression

Cut 5 intentionally changes one diagnostic path.

The failure golds established at Cut 0 make that movement explicit. Parsing results must not change with the diagnostic.

### Scope Creep into Language Design

The grammar analysis may make the case for leading `|>`, `.`, or binary operators compelling.

Do not change them here.

A behavior-preserving infrastructure proposal must not quietly become a syntax redesign. Any such change requires a separate proposal, separate behavior-gold changes, and an explicit language decision.

### Coverage Bureaucracy

Bidirectional checks are valuable only when they protect a real coupling.

Do not force formatter, linter, evaluator, and parser into superficially identical dispatch architectures merely to make set equality convenient. Each invariant should match the subsystem's real responsibility.

The goal is **detectable drift**, not architectural symmetry for its own sake.

## Open Questions

### 1. Should the Newline Rule Later Be Narrowed?

If the grammar confirms that only:

```text
(  [  -
```

are both statement starters and expression continuations, Aiki could potentially permit leading:

```text
|>  .  +  *  /  <  >  <=  >=  and  or
```

without introducing statement-boundary ambiguity.

That is a language-design decision, not part of this refactor.

The derived `overblocked continuations` set should provide the evidence for a later proposal.

### 2. Should `@newline` Remain Aiki-Specific?

The proposed block encodes a particular normalization model: preceding-token completion plus delimiter suppression.

Another language using the EBNF machinery might want indentation, explicit newline terminals, different delimiter rules, or context-sensitive layout.

That is acceptable for now. `grammar.ebnfx` is first the specification machinery for Aiki. Generality should be earned by a second use case rather than anticipated abstractly.

### 3. Is `;` in `program` and `block` Still an Honest Representation?

After the rule becomes explicit, the relationship becomes clearer:

```text
program = { statement [ ";" ] }
block   = "{" { statement [ ";" ] } "}"
```

describes the token stream consumed by the grammar, while newline normalization may synthesize those terminators before parsing.

That may be perfectly honest once documented. If not, the distinction between explicit separator and synthetic terminator deserves a later grammar-design discussion.

Do not alter these productions merely to make the notation philosophically tidier.

### 4. What Exactly Does the REPL Treat as a Parse Unit?

The investigation suggests that the REPL may parse one entered line at a time, which would make multi-line pipelines only one instance of a broader interactive limitation.

That must be established observationally.

It is out of scope for this proposal and should receive its own investigation if confirmed.

## Why This Fits Aiki

Aiki is intended to make computational relationships visible enough to inspect, reason about, and receive.

A rule governing something as ordinary as where a programmer may break a line should therefore not be discoverable only by violating it and then reading the Go implementation.

The project has repeatedly moved toward the same architecture:

> locate the artifact that should be authoritative, make it authoritative, and mechanically guard the places where another subsystem depends on it.

`grammar.ebnfx` did that for syntax. Executable documentation does it for documented behavior. Structural and behavioral golds do it for cross-artifact claims. `treecheck` does it for the distribution tree.

This proposal completes an unfinished part of the grammar boundary.

The defect is not that leading `|>` currently fails. That may remain exactly the right Aiki rule.

The defect is that the grammar appears to permit a surface form, token normalization silently forbids it, and the system contains no mechanism capable of identifying the disagreement.

**Aiki may choose restrictive syntax. It should not acquire accidental syntax.**
