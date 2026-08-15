# Cut 1 — Declare negative fixtures

Status: **GATED (environment-limited)**

## Intent

Give intentionally failing smoke specimens an explicit declaration of intent
rather than inferring intent from their gold transcripts.

## Result

- Added the textual declaration `# @negative parse`.
- Added a shared declaration reader under `cmd/internal/testfixture`.
- Only `parse` is supported initially; unknown or duplicate declarations fail.
- Marked the three existing leading-continuation parser-negative specimens.
- Smoke validates declaration versus parser-failure observation in both
  directions, both during ordinary comparison and before blessing.
- fmt and lint skip declared parse-negative specimens.
- Documented the convention in `docs/testing.md`.

The leading marker shifts source line numbers by one; the three affected golds
were corrected only for that source-position movement.

## Evidence

Local unit tests passed for the declaration reader, smoke coupling, and formatter
package using the available Go 1.23 compatibility harness.
