# Milestone 01 — Public README and project-method framing

Status: GATED

## Intent

Replace conference-specific/project-internal landing-page framing with a truthful public alpha entry point, while making the project's unusual AI-assisted engineering provenance explicit.

## Changes

- Removed the live SPLASH-E release sentence from the README and durable AI trigger text.
- Reframed the repository as an Aiki alpha release.
- Identified Linux as the developed/tested/supported alpha target.
- Characterized macOS and Windows binaries as untested, advanced-user builds.
- Added Getting Started instructions that explain why Go and Ebitengine appear in an Aiki source build.
- Made `make validate` the primary source-build confidence path before running `./aiki`.
- Simplified the README documentation/development path.
- Added an `AI and Authorship` section defining the distinction between design authority and generated implementation labor.
- Added an `Invariant Framework` section explaining the executable couplings and gates that enforce that distinction.
- Clarified BSD 3-Clause licensing and copyright in the README.

## Design decisions

### Alpha versus public readiness

`alpha` describes language maturity. Public GitHub readiness is a separate quality bar: a newcomer must be able to understand what Aiki is, build or run it, know which platform is supported, and understand the status and provenance of the project without being misled.

### AI authorship boundary

Generative AI may propose, criticize, and produce implementation artifacts. It does not acquire design authority. Suggestions become part of Aiki only through the designer's judgment. Correctness is established against externally imposed specifications, behavior, executable documentation, tests, gold files, structural invariants, and validation gates.

### README role

The README is the front door, not the complete manual. It should orient a new user, produce a first successful run, point toward the curated documentation, and state the project's method and provenance clearly.

## Evidence

- README contains no live SPLASH-E framing.
- Existing project validation machinery is described rather than replaced.
- README points to existing repository paths and verified sample/module surfaces.

## Caveat

Full repository validation remains an end-of-session release gate and was not claimed by this milestone.
