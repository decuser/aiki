# Milestone 19 — Dynamic value construction gap

Status: SUPERSEDED — resolved by Milestone 20 with `to_symbol` and `shaped`.

## Finding

The independent evaluator must construct runtime values from source lexemes.
Ordinary user-level Aiki can already construct several required values
dynamically:

- numbers via `to_number(string)`;
- runes via `chr(number)`;
- strings through ordinary string/rune operations;
- booleans and lists through ordinary language operations.

It cannot, however, construct an arbitrary native symbol from a string. There
is no user-visible string-to-symbol operation in the current prelude/HAL
surface. A source literal such as `:arbitrary_name` therefore cannot be turned
into the corresponding native Aiki symbol without enumerating symbol names.

Arbitrary shaped-list construction has the same issue. The substrate already
contains `_make_shaped_list(symbol, elements)` and native modules use it, but it
is HAL-only and therefore unavailable to ordinary user-scope Aiki. Moreover,
it still requires a native symbol for the dynamic shape name.

## Why this is load-bearing

Hard-coding a finite symbol table would make the self-hosted interpreter cease
to be an implementation of the language's open symbol syntax. Representing
symbols/shapes as private wrapper values would abandon the proposal's stronger
claim that interpreted Aiki values are ordinary Aiki values and would complicate
or break the later prelude/module bridge.

Generating a temporary source file and calling `load()` is also invalid as a
solution: `load()` lexes, parses, and evaluates through the host Go interpreter,
so using it to manufacture source literals would silently delegate the very
semantics the second implementation is meant to prove.

## Evidence

`prelude.help` exposes conversion/inspection operations including `to_str`,
`to_number`, `ord`, `chr`, `type`, `shape`, and `inspect`, but no general
string-to-symbol constructor. The HAL registry contains `_make_shaped_list`,
which confirms shaped-list construction exists below the user/prelude surface
but is not an ordinary user-level capability.

## Architectural interpretation

This is exactly the kind of gap the self-hosting proposal said should be
surfaced rather than worked around. It is evidence about the sufficiency of the
current platform/language boundary, not a defect in the independent lexer or
parser.

A minimal general-purpose resolution would need to be evaluated on language
merit, not introduced as a hidden self-host exception. Candidate surface:

- a general string -> symbol constructor (for example `to_symbol(string)`);
- a general shaped-list constructor that takes a symbol/name plus elements, or
  an equivalent language-level operation.

The exact names and surface are deliberately not committed in this milestone.

## Next action

Review and decide the general language surface for dynamic symbol and shaped
list construction. Once that capability is deliberately available to ordinary
Aiki, resume Cut III.1 and use it through the same public/prelude-level surface
as any other Aiki program.
