# Store checksum boundary pass

The byte-grid coordinator boundary removed interpreted Store-to-text rendering from
`grid_text`, but profiling then exposed `life.store_checksum` as the next
coordinator-side scan:

- one interpreted `store.get` per cell
- one interpreted `mod` per cell
- one interpreted `truncate` path through rational arithmetic per cell

This pass adds `store.checksum(store[, count])` as a coarse Store primitive using
the same diagnostic formula:

```
value = 17
value = ((value * 131) + cell + index) mod 1000000007
```

The coordinator now computes status checksums directly at the Store boundary.
Worker byte-grid computation is unchanged.
