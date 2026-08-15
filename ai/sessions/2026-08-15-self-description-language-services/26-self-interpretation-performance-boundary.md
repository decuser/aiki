# Milestone 26 — Self-interpretation performance boundary

Status: SUPERSEDED — subsequent measurement and correction completed the full proof; see Milestone 27.

## Attempt

The outer Aiki-written interpreter was asked to evaluate an inner interpreter bootstrap that imports the independent lexer, normalizer, parser, evaluator, runtime, and host-prelude bridge, then evaluates a small program (`1 + 2 * 3`).

A traced run reached:

```text
N0  inner bootstrap entered
N1  inner lexer loaded
N2  inner normalizer loaded
```

It then remained in the outer self-host parser while parsing/evaluating `selfhost/parser.ai` for more than one minute; a longer attempt exceeded two minutes. No semantic fault or disagreement was emitted.

## Interpretation

This is a bootstrap-performance boundary, not evidence that self-interpretation is semantically impossible. The independent parser is sufficiently fast when executed by the Go bootstrap, but parser-self-parse under a tree-walking interpreter compounds environment lookup, AST traversal, and recursive-descent costs substantially.

## Constraint

Do not weaken the proof by silently host-parsing the inner parser or sharing Go parser structures. Any next step must preserve independent implementation.

## Next options

1. profile the nested parser-self-parse using Aiki's existing semantic instrumentation and optimize implementation-neutral hot paths;
2. define a smaller but still honest self-interpretation proof kernel and state precisely what it proves;
3. accept a deliberately long-running full proof outside ordinary `make validate` if measurement shows finite completion.

The next session should measure before choosing among these.


## Follow-up

The original diagnosis was refined by measurement. The first dominant cost was repeated whole-string rune conversion in the independent lexer. After that was corrected, the trustworthy nested proof exposed two module-resolution issues: missing current-directory fallback for path imports and spelling-sensitive cache keys for paths containing `.`/`..`. Once path behavior matched the host loader and module identities were canonicalized, the full nested proof completed successfully. See `27-self-interpretation-complete.md`.
