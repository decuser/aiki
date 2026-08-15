# Summary: Post-Grammar Exceptional-Path Hardening

This project arose from the combined post-branch code audits after grammar
surface authority and centralized grammar analysis were completed. Rather than
turn every review comment into immediate work, the project first created a
repository-wide findings ledger with stable `AF-###` identifiers and explicit
`RESOLVED`, `ACCEPTED`, and `DEFERRED` dispositions.

The bounded implementation addressed exceptional paths where hidden or overly
broad state could undermine the new authority model. Parser newline suppression
previously used one aggregate depth, so a stray closer could drive the depth
negative and cancel a later legitimate opener. It now tracks expected closing
delimiters, ignoring unmatched/mismatched closers for normalization state while
leaving the grammar parser responsible for the malformed syntax itself.

Cached newline analysis remains an optional diagnostic refinement rather than a
parse prerequisite. When derivation is unavailable, parser construction still
succeeds, but observers that implement the optional diagnostic extension now
receive the reason instead of experiencing silent degradation.

Negative-fixture intent is now explicitly test-scoped. `# @negative parse` is
valid only on `*_smoke.ai` fixtures; the same marker in ordinary source is an
error and cannot exempt application/library source from fmt or lint. Lint path
expansion was simplified so one candidate-classification helper owns its source
and negative-fixture disposition instead of repeating the logic across three
branches.

The closeout also documented the EBNF scanner's current ASCII identifier
assumption, cross-referenced parser-synthesized `TERMINAL` from evaluator
coverage policy, and explained the two intentionally empty ambiguity smoke
golds. Other audit findings remain explicitly accepted or deferred in
`docs/audit-findings.md` rather than disappearing from project memory.
