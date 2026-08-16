# Milestone 13 — HAL redesign Cut 0

Status: **GATED**

Baseline: `v0.4.0-alpha-27`.

The exploratory HAL discussion was consolidated into
`proposals/hal-redesign-cut0.md`.

The design premise is explicitly user-facing: the redesign exists to cleanly
realize the host-facing affordances an Aiki systems programmer needs. It is not
an architecture cleanup exercise for its own sake.

The standing three-name distinction is load-bearing:

```text
Aiki name       = programmer meaning
HAL name        = architectural contract
Substrate name  = implementation realization/provenance
```

Canvas is retained as a pressuring requirement, not the driver.

The design effort is structured as Cut 0 plus three phases of four cuts each
(13 cuts total). Phase I begins with source-derived inventory/classification.
