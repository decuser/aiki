# Milestone 03 — Publication hygiene and historical scratch scrub

Status: GATED

## Intent

Remove low-value private-development debris before public release and verify that the reachable repository does not expose secrets or accidental local-machine information.

## History cleanup

The historical tracked files `out` and `output` were debugging debris. Their contents included local `/home/wsenn/...` paths and, in one case, author metadata copied into diagnostic output. They had no continuing design or provenance value.

Both paths were removed from every reachable ref by rewriting repository history. The rewrite preserved the canonical history, tags, and historical side branches while changing commit IDs for descendants of affected commits as expected.

## Public/sensitive-information scan

The current tree and reachable history were checked for:

- private-key markers;
- common API/token/credential forms;
- password/secret assignments;
- credentialed URLs;
- personal home-directory paths;
- email addresses in tracked content;
- obsolete live SPLASH-E framing;
- archive/binary debris and anomalously large reachable blobs.

### Findings

- No credentials, private keys, API tokens, passwords, or credentialed URLs were found.
- No personal home paths remain in reachable tracked content.
- No email address remains in reachable tracked content.
- Commit author/committer metadata intentionally identifies `Will Senn <will.senn@gmail.com>` as historical authorship provenance; this was retained by design.
- The only tracked office-document binaries are the intentional language documents `docs/ra1.odt` and `docs/this-is-aiki-formatted.odt`.
- No reachable blob is 100 KiB or larger after the scrub.
- Live repository content contains no SPLASH-E framing. Historical snapshots may retain old conference-specific wording as archaeology; that history was intentionally not rewritten for wording alone.

## Integrity evidence

- `git rev-list --objects --all` reports zero reachable objects at paths `out` or `output`.
- `refs/original/` and `refs/replace/` are absent after cleanup.
- `git fsck --full` found no reachable integrity error.
- Current release-prep working-tree content was preserved across the history rewrite.

## Consequence

Because history changed, the private/local remote must be force-updated before the repository is published elsewhere. This is intentional and safe while the repaired alpha history remains unpublished.
