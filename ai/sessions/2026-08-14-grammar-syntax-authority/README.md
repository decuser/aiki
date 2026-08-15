# Grammar as sole syntax authority — 2026-08-14

Status: COMPLETE

## Intent

Execute `proposals/grammar-as-sole-syntax-authority-redux.md` so that
`grammar.ebnfx` becomes the sole authority for Aiki surface syntax policy at the
newline boundary and grammar-sensitive implementation couplings become
mechanically checked.

## Baseline

- Branch: `grammar/newline-rule`
- Baseline commit: `f9e53c7`
- Supplied baseline: `aiki(20260815-032602).tgz`
- Working tree at start: clean
- Prior session: `../2026-08-14-alpha-release-prep/` (COMPLETE)

## Result

The proposal was implemented in nine gated milestones: Cut 0, Cuts 1–3,
Cuts 4a–4d, and Cut 5. The final authoritative `make validate` passed after the
Cut 5 grammar structural gold was regenerated with Aiki's own `enginesmoke`
blessing command.

Final validation evidence:

- all Go tests passed;
- 408 Aiki-native tests passed;
- behavior smoke: `smoke ok (46 tests)`;
- grammar coverage: 32 productions across 10 inputs;
- engine structural golds: 10 inputs passed;
- treecheck: 471 files, 415 structurally justified, 56 explicitly allowed.

## What is true now

- `grammar.ebnfx` explicitly declares the newline token, completion token types,
  completion lexemes, suppression pairs, and user-facing help.
- `parser.go` consumes that declaration; `isComplete` and the private Aiki
  newline vocabulary are gone.
- grammar analysis derives expression endings, statement starts,
  continuations, ambiguous continuations, overblocked continuations, uncovered
  expression endings, and impossible declaration entries.
- evaluator handler coverage is bidirectional over the real AST-producing set.
- formatter coverage is explicit across all 32 productions, including six
  parent-handled productions, and unknown leaves cannot silently disappear.
- linter grammar-sensitive node names are checked while generic traversal is
  preserved.
- `BINOP` membership comes from the grammar while evaluator code remains the
  authority on operator meaning; the coupling is bidirectionally checked.
- `help("newline")` comes from grammar metadata.
- newline-caused leftover-token diagnostics explain the grammar-declared rule
  without a private continuation list.

## Discoveries

1. The ambiguous followers `(`, `[`, and `-` do not fail after newline in the
   baseline. Synthetic termination makes them new expression statements.
2. The grammar derives overblocked unambiguous continuations:
   `* + . / < <= > >= and or |>`.
3. `}` can end an expression but is absent from the newline completion set. A
   function literal can therefore continue across a newline. Current behavior
   is pinned by `newline_function_end_smoke.ai` and intentionally preserved.
4. Recursive `fmt ./...` and `lint ./...` traverse intentionally invalid
   negative smoke fixtures and print their expected parse diagnostics. This is
   deferred as `buglist.md` B1.

## Decisions

`docs/decisions.md` D2 records the decision not to turn the Cut 3 analysis into
a syntax change. This proposal establishes authority and detectable drift; it
does not redesign newline policy.

## Next action

No action is required to complete this proposal. A separate future proposal may
address newline policy (`}` termination and/or unambiguous leading
continuations). Separately, B1 may receive a tooling cut to make recursive
fmt/lint fixture-aware.
