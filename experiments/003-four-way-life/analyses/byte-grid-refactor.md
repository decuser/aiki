# Byte-grid refactor cut

This cut changes the Four-Way Life worker-side grid representation from an
immutable Aiki list reconstructed by per-character string indexing and repeated
`append` into a dense byte grid.

The line protocol remains textual because workers communicate over line-oriented
pipes. At the boundary, `bytes/ffi.digits_from_text` decodes ASCII digit text into
raw byte values `0..9`; `digits_to_text` performs the inverse. After decoding,
Life rules access cells through `bytes.bytes_get` rather than list indexing over a
freshly appended list.

The goal is not to optimize strings. The goal is to treat the Life grid as dense
memory while preserving the current process protocol shape.

Expected profile effect:

- remove the `grid_from_text` repeated-append hotspot
- reduce `evalStringIndex` contribution from grid reconstruction
- leave ordinary `append` semantics unchanged
- keep `string/ffi` responsible only for protocol field splitting/joining

