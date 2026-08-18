# Byte-grid direct access pass

The first byte-grid cut removed the string-to-list reconstruction hotspot, but
profiling showed that the worker hot loops still paid a generic `life.cell`
dispatch cost on every cell access:

```
cell -> type -> equal -> bytes_get
```

This pass keeps the generic `cell`/`grid_length` helpers for compatibility, but
adds byte-specific hot-path functions and routes worker computation through
them. The line protocol remains text at the process boundary; after decoding,
the worker grid is treated as byte-backed memory.

Expected profile effect: `life.ai:cell -> type/equal/bytes_get` should disappear
from the top attribution, replaced by direct `bytes.bytes_get` calls at the Life
algorithm sites.
