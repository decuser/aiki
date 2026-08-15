# Milestone 03 — Negative Fixture Boundary

**Status:** GATED (focused validation)

Resolved AF-016 by making `# @negative parse` valid only on `*_smoke.ai`
fixtures. A marker in ordinary `.ai` source is an error, so it cannot silently
remove source from fmt/lint coverage.

Resolved AF-014 by rewriting lint path expansion around one
`lintSourceCandidate` disposition helper and one directory-walk helper rather
than repeating negative-fixture classification in three branches.

Updated fmt/lint tests to use correctly named smoke fixtures and added negative
scope tests. `cmd/internal/testfixture` passes in the local compatibility test.
Lint/fmt full package execution remains environment-limited by unavailable
Ebitengine dependencies and is deferred to the authoritative final gate.
