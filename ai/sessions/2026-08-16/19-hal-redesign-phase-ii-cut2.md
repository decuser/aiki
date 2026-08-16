# Milestone 19 — HAL redesign Phase II / Cut II.2

Status: **GATED**

Defined the conceptual runtime/session ownership split. Runtime owns host
relationships, module/cache identity, args/I/O/environment view, RNG,
observation correlation, async faults, capabilities, and host resources.
Evaluation session owns evaluator interaction, prelude/user environments, and
presentation/dynamic evaluation state.

Identified core-value substrate leakage: `value.File` embeds `*os.File` and
`value.Canvas` embeds Go graphics/channels/protocol state. Future host-resource
values should not import substrate realization into the semantic value layer.
