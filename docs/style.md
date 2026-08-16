# Aiki Project Source Style

Aiki uses two formatting layers with distinct responsibilities:

```text
go fmt / aiki fmt   canonical language formatting
aiki distfmt        project distribution/source presentation
```

`fmt` remains the canonical formatter. `distfmt` applies this repository's
presentation style only through layouts that the canonical formatters preserve.
The Make `fmt` target runs canonical formatting first and `distfmt` last; every
Make workflow that depends on `fmt` therefore leaves the tree in project style.

## Governing rule

Use compact source for ordinary expressions and declarations. Expand long,
primarily declarative structures when doing so makes them materially easier to
audit. `distfmt` uses one fixed project width and rewrites only syntactic forms
whose structure it can prove unchanged.

Current rules:

- long Go elided composite literals used as map values are expanded vertically;
- long Aiki lists and calls are expanded vertically;
- compact forms remain compact;
- there is no user-configurable width or alternate style mode.

Example Go policy entry:

```go
"lib/bits/bits.ai": {
    "_bits_and",
    "_bits_not",
    "_bits_or",
    "_bits_shl",
    "_bits_shr",
    "_bits_xor",
},
```

`gofmt` preserves that explicit multiline composite literal. Aiki follows the
same project principle without changing the grammar: when a list, call, or
function parameter list is explicitly multiline, `aiki fmt` normalizes and
preserves the multiline form rather than collapsing it. Compact input remains
compact. `distfmt` therefore chooses where project style expands source, while
`aiki fmt` guarantees that the chosen layout is stable afterward.

## Safety

`aiki distfmt` follows the operational model of `aiki fmt`:

- hidden directories and parse-negative Aiki fixtures are skipped during
  recursive traversal;
- source is canonically parsed/formatted before restyling;
- restyled output is parsed again and must be structurally equivalent;
- Aiki restyling is passed through `aiki fmt` again so the result is already a
  canonical, formatter-stable expanded form;
- writes use a temporary file, `fsync`, and rename rather than in-place partial
  writes;
- `-n`, `-p`, and `-b` mirror `aiki fmt` behavior.

A presentation rule that cannot be proved structure-preserving does not write.
