# 02 — Authority and representation inventory

Status: GATED

## Intent

Complete Phase I, Cut I.1 before implementing the independent lexer. The goal
is to identify the actual baseline authorities and distinguish language facts
from Go implementation representation.

## Lexical authority

`engine/syntax/grammar.ebnfx` is the syntax authority. Its `@tokens` block
currently declares twelve token classes in order:

```text
WHITESPACE NEWLINE COMMENT NUMBER STRING RUNE SYMBOL SHAPE
KEYWORD OPERATOR DELIMITER NAME
```

Pattern token definitions own the lexical regular languages. Literal token
definitions own keyword/operator/delimiter inventories. `@skip` is grammar
metadata and applies to WHITESPACE and COMMENT. The Go lexer consumes
`grammar.Grammar.Tokens`; it does not own separate production vocabulary.

Cut I.0 can already derive and check the ordered token names and literal
keyword/operator/delimiter inventories. Pattern behavior remains independently
implemented and is proved by token conformance fixtures rather than by sharing
Go regexp machinery.

## Go token representation

`engine/syntax.Token` contains:

```text
Type   string
Lexeme string
Pos    engine.Position
```

`engine.Position` contains file, one-based line, and one-based column. EOF is a
Go lexer sentinel returned by `Next`; `Tokenize` does not include EOF in its
result. The lexer emits @skip tokens as ordinary tokens; parser normalization
removes them.

The initial neutral token conformance projection can therefore be exactly the
language-observable subset:

```text
kind | lexeme | line | column
```

The file name is harness context, not lexical behavior, and should not be in
fixture equality unless a later diagnostic contract requires it.

## Lexical matching policy

The Go lexer is grammar-driven and implements maximal munch. Ties are resolved
by earlier grammar token definition, then earlier literal within a literal
token definition. This is observable behavior that the independent Aiki lexer
must reproduce; it is not an authority to copy blindly. Fixtures must cover
prefix/tie cases such as `<` versus `<=`, `|`-prefixed operators, names versus
keywords, and literal forms.

## Newline authority and normalization

The `@newline` block in `grammar.ebnfx` is the sole policy authority. It owns:

- physical newline token name;
- `after_token` completion classes;
- `after_lexeme` completion literals;
- delimiter suppression pairs.

The Go parser compiles this policy in `NewParser`, filters `@skip` tokens, and
turns eligible physical newlines into synthetic `DELIMITER ;` tokens. It uses a
closer stack rather than a scalar depth, so mismatched/unmatched closers do not
corrupt suppression state. Grammar analysis derives expression continuation
sets only for diagnostics; those sets do not define normalization policy.

Phase-I newline fixtures should project the filtered stream using the same
neutral token representation and therefore make synthetic semicolons visible.

## Syntax representation

`engine/syntax.Node` contains:

```text
Type     string
Value    string
Children []*Node
Pos      engine.Position
```

`Type` is a production name or token type; terminals carry their lexeme in
`Value`; nonterminals carry ordered children. Convenience methods such as
`ChildByType` are Go implementation API and are not part of conformance.

This representation is already close to a neutral syntax projection, but
Phase I will not canonize the Go struct. Candidate fixture form is a recursive
language-owned projection containing:

```text
kind
value when terminal/significant
line
column
children in grammar order
```

Whether all production wrapper nodes and all positions are significant will be
decided against the independently implemented parser rather than assumed from
Go structs.

## Existing authority/invariant machinery

Relevant baseline machinery includes:

- grammar loader/types under `engine/syntax/grammar/`;
- cached central `Grammar.Analysis()` structural knowledge;
- newline structural analysis derived from grammar productions and policy;
- grammar/lexer/parser unit tests;
- engine structure/gold tests and grammar coverage;
- engine observers for lex/parse events;
- Cut I.0's new Aiki-native lexical-authority projection.

## Duplicated facts requiring coupling

The independent implementation may necessarily duplicate:

- keyword/operator/delimiter tables;
- pattern-recognition logic;
- newline completion/suppression tables;
- hand-written parser production structure.

Enumerated grammar facts must be checked against the grammar authority.
Algorithms remain independent. Parser structure is checked by reviewed syntax
fixtures and production coverage rather than generated from Go code.

## Gate conclusion

The baseline authorities, observable Go representations, newline policy, and
candidate neutral token/syntax projections are now identified sufficiently to
begin the independent lexical scanner without inventing a second authority.
No implementation change was required for this inventory.

### Follow-up: position-unit discovery

While implementing the independent lexer, source-position semantics were found
to be representation-sensitive for non-ASCII input. The Go lexer increments its
column once per UTF-8 byte; Aiki string indexing advances by rune. ASCII input is
identical. No opportunistic semantic change was made. Phase-I conformance is
restricted to ASCII positions and Phase II must decide the language-level
column unit before LSP/editor translation.
