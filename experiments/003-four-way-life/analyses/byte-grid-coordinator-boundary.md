# Byte-grid coordinator boundary pass

The worker byte-grid passes removed the hot worker-side `grid_from_text -> append`
and then removed generic `cell -> type/equal` dispatch from worker computation.

The remaining high-volume coordinator path was not worker grid computation. It was
the coordinator's transport boundary:

- `grid_text` walked the Store one cell at a time
- each cell called `store.get`
- each value called `to_str`
- each digit was appended by string concatenation

This pass adds `store.digits_to_text(store[, count])` as a coarse Store-to-text
boundary for decimal digit stores and uses it for the Four-Way protocol frame.
It also adds `life.store_checksum(store, count)` so the coordinator does not
round-trip Store -> text -> bytes just to compute the status checksum.

This remains representation-faithful: the coordinator owns mutable Store memory;
workers receive dense byte grids; the protocol remains line-oriented text.
