# String observation witness note

String indexing returns an Aiki `rune`, not a one-character `string`.
Accordingly, equality witnesses must compare indexed results with another rune,
for example `first("猫")`, rather than with `"猫"` directly.

This distinction is semantic and intentional:

```text
"猫"          -> string
first("猫")   -> rune
text[i]       -> rune
```
