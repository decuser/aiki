# Milestone 18 — HAL redesign Phase II / Cut II.1

Status: **GATED**

Decomposed current `Env`/`EvalContext` responsibilities into lexical, source,
dynamic, observation, authority, and module-evaluation facets. Traced ordinary
call, spawn, module evaluation, and native invocation.

Important baseline finding: spawned user functions are created under the
prelude environment and inherit `ScopePrelude`; builtin lookup is scope-based.
This couples spawn isolation to raw native authority and is a source-derived
authority leak that the redesign must remove.
