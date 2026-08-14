# 02 - First-class profile module

Status: **GATED**

## Intent

Make semantic measurement available to Aiki programs as ordinary Aiki operations and ordinary data.

## Implemented

`lib/profile` provides:

- `counts(function)`;
- `experiment(size, function)`;
- `measure(function)`;
- `complexity(results)`.

Help, documentation, and module tests accompany the surface.

`complexity` is implemented in Aiki rather than Go. It interprets empirical experiment results and reports the best observed growth fit; it does not claim that finite measurements prove asymptotic complexity.

## Semantics pinned down

Tests establish that:

- setup performed before the measured function is excluded;
- each experiment size receives fresh counters;
- ordinary program state is not magically reset between samples;
- complexity analysis requires at least three sizes.

The positional result representation was retained during this cut rather than mixing API redesign into validation.

## Validation

- focused `lib/profile` tests - 20/20 PASS;
- full Aiki-native suite - 406/406 PASS.

## Next step established

Add opt-in attribution so counts can answer where semantic work occurred in source, not merely how much work occurred in total.
