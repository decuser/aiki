# Milestone 01 — Cut 0 baseline and probes

Status: GATED

Established the observational baseline before parser changes. Eleven behavior
smokes pinned trailing/leading pipeline behavior, access and binary operator
continuations, delimiter suppression, ambiguous `(`/`[`/`-` cases, blank
lines, trailing comments, and block context.

The first local gate was environment-blocked by Go 1.24 toolchain acquisition,
so no claim was made from the disposable Go 1.23 harness. The user's
Go-1.24-capable environment then ran `make validate` successfully after the
new negative-smoke golds were corrected to the smoke transcript format.

Discovery: `break_call`, `break_index`, and `break_unary` pass in the baseline.
Termination resolves those ambiguous followers as new expression statements.
The proposal was corrected before implementation proceeded.
