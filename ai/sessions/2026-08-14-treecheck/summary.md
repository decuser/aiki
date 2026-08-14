# Summary — treecheck

Aiki's overlay-based development workflow makes additions and modifications
cheap, but it can leave an old path behind after a rename or deletion. Rather
than require historical comparison or a complete distribution manifest, this
session generalizes Aiki's existing disposition invariants to the repository
itself: every file should either participate in a recognized structural
relationship or be explicitly allowed as a standalone artifact.

The intended result is `aiki treecheck`, wired into validation. A small root
allow file handles intentional standalone material; it is an exception list,
not an inventory of the distribution.

The implemented checker proved useful immediately: its first pass found eight
real stale artifacts without producing broad false positives. The resulting
design is deliberately asymmetric: infer common relationships aggressively,
but keep intentional standalone material explicit in a small exception file.
This preserves the simplicity of overlay-based delivery while making orphaned
paths observable during validation.

