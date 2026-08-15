# Milestone 06 — Cut 4b formatter coupling

Status: GATED

Made formatter production coverage explicit and introspectable. All 32 grammar
productions have a disposition: 26 explicit formatter dispatches and six
parent-handled productions (`field`, `select_case`, `select_default`,
`param_list`, `rest_param`, `literal`).

The previous silent default recursion path was removed for unknown nodes; an
unknown leaf now produces an explicit formatter error rather than disappearing.
A regression test pins that protection.

An attempted method-closure dispatch map caused a Go initialization cycle and
was corrected before handoff to a production-to-handler-kind table.

Authoritative `make validate`: passed.
