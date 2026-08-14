# Milestone 09 - next priorities

Status: **PLANNED**

With the profiling work complete and the semantic items formerly in `buglist.md` resolved, the next work should proceed in these serial cuts:

1. Remove the `value -> engine` profiling dependency.
2. Clean documentation and buglist drift.
3. Re-run `validate` and the executable-coupling checks.
4. Revisit spawned abnormal termination carefully as a design question, without assuming joins, futures, task handles, or channel close.
5. Add mundane systems substrate capabilities:
   - program arguments and environment access;
   - directory listing;
   - random-access file operations.
6. Consider extracting a language-services layer before building editor or notebook adapters.

## Restart point

Begin with item 1 only. Preserve the single authoritative tree, take a small cut, record findings here or in the next milestone, and gate it before moving to item 2.
