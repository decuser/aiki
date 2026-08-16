# Milestone 15 — HAL redesign Phase I / Cut I.2

Status: **GATED**

Classified the current native registry into:

- true host effects/resources;
- language/evaluator intrinsics;
- language/value primitives implemented natively;
- optional library accelerators/FFI realizations;
- observation/tooling/runtime services;
- context/ownership concerns adjacent to callable machinery.

Decision: a replacement HAL must not simply re-express the current 117-entry
registry as 117 canonical host operations. Native does not imply host-backed.
