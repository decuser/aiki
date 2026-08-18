# String FFI result

Four-Way Life exposed pure-Aiki string parsing as a generic runtime/library
hotspot. The explicit `string/ffi` experiment kept bare `string` unchanged and
switched the experiment protocol/worker parsing to the accelerated sibling.

Against the showcased `58c7b64` baseline, representative one-generation worker
profiles improved materially:

- Worker 1 elapsed: about 183 ms -> 129 ms (about 30% faster).
- Worker 4 elapsed: about 189 ms -> 158 ms (about 16% faster).
- Worker 1 allocation: about 213 MB -> 145 MB.
- Worker 1 mallocs: about 1.74 M -> 1.29 M.
- Worker 4 allocation: about 220 MB -> 152 MB.
- Worker 4 mallocs: about 1.90 M -> 1.44 M.

The result supports Aiki's existing native/FFI pattern: coarse provider calls
that replace substantial interpreted work can amortize the boundary. A prior
experiment using `bytes/ffi` for fine-grained grid indexing did not improve
wall time, reinforcing that FFI crossings should be coarse.

After string acceleration, allocation attribution exposed `_append` as the
next steady-state generic hotspot, while parser failure recording remains a
separate process-startup hotspot.
