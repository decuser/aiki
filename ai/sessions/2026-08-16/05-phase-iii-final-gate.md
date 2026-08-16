# Phase III final gate and reconstruction completion

Status: GATED

## Authoritative local run

The corrected experiment runner was executed under:

```text
Aiki executable: /home/wsenn/forge/dev/aiki/aiki
Aiki version:    aiki v0.4.0-alpha-26
```

Retained result:

```text
experiments/002-thompson-7094-regex/results/run-2026-08-16-001438.448.txt
```

## Gate results

```text
Phase I machine corpus:       96 tests, 96 passed, 0 failed
Phase II Thompson corpus:     59 tests, 59 passed, 0 failed
Phase III compiler corpus:    39 tests, 39 passed, 0 failed
```

The human-readable end-to-end demonstration then completed successfully.

## Demonstrated path

```text
a(b|c)*d
  -> Stage 1: a.(b|c)*.d
  -> Stage 2: abc|*.d.
  -> Stage 3: 23 IBM 7094 words
  -> Aiki IBM 7094 emulator
  -> Thompson runtime
  -> FOUND signals
```

Representative generated-code searches were:

```text
"ad"          -> FOUND end offsets [2]
"abd"         -> FOUND end offsets [3]
"abcbcd"      -> FOUND end offsets [6]
"xxabcbcdyyad" -> FOUND end offsets [8, 12]
"axd"         -> no FOUND offsets
```

The generated Stage-3 object program independently produces `TRA CODE+16` at word 4, while Thompson's printed example lists `TRA CODE+13`; the discrepancy is therefore preserved as a source finding rather than silently normalized.

## Conclusion

The baseline reconstruction is complete and genuinely gated under the corrected Aiki test runner. Earlier nominal GATED/COMPLETE claims based on ordinary script execution remain superseded by this run.

No extension work (lambda/a** correction, redundant-search suppression, broader 7094 coverage) is part of this baseline completion.
