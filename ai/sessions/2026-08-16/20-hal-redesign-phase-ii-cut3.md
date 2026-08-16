# Milestone 20 — HAL redesign Phase II / Cut II.3

Status: **GATED**

Settled the working authority model: HAL grants are explicit, lexical to trusted
Aiki definitions, domain/operation-scoped, runtime-policy assigned, and
independent of filesystem topology. User modules receive no raw host grants;
public Aiki library closures carry the authority needed to realize their APIs.
Prelude/evaluator privilege is separated from host authority.
