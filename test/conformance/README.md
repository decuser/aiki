# Conformance Tests

Verify that independent implementations and distribution integrations preserve
Aiki behavior.

These tests are behavioral evidence and run under `make test`, not
`make invariant`. They may support architectural claims, but they execute the
language to establish conformance rather than inspect architectural structure.

The self-host suite lives here because it compares the independent Aiki-written
front end/evaluator and self-interpretation behavior against the reference
implementation and reviewed corpora.
