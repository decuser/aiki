# Per-session AI records are retired

Do not create per-session directories under `ai/sessions/`.

Durable execution state belongs in:

```text
ai/README.md               working method
proposals/                 bounded design contracts
docs/session-history.md    concise cumulative execution history
```

Chat transcripts, cut-by-cut narration, temporary handoff notes, and generated
session folders are repository cruft. Before the next release tag, run
`make historycheck`; post-tag session/proposal cruft should be removed from the
working tree and, when the affected history is still private, excised from the
post-tag history before it is pushed or tagged.
