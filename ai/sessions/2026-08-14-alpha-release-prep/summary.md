# Alpha release prep summary

Aiki is being prepared for public GitHub release as an alpha. This session concentrates on the public artifact rather than adding language features.

The README now treats Aiki as the project itself rather than as an artifact attached to SPLASH-E. It provides a Linux-first path from prerequisites to `make validate` to `./aiki`, explains Go and Ebitengine as implementation/runtime prerequisites rather than user-facing language concepts, and explicitly characterizes macOS and Windows builds as untested advanced-user artifacts.

The README states alpha limitations, a deliberately generic roadmap, forthcoming introductory books aimed in part at non-programmers and younger learners, a verified `turtle/simple` example, and an invitation for useful alpha feedback.

It also records the project's authorship boundary. Will Senn (decuser) retains design authority over the language, architecture, semantics, constraints, and acceptance criteria. Generative AI has been used extensively to produce implementation and supporting engineering material under that authority. AI suggestions may enter the design conversation, but generated output has no authority merely because it was generated.

That boundary is operationalized by Aiki's invariant framework: behavioral tests, gold files, executable documentation, grammar-handler coverage, formatter/AST coverage, module/export checks, documentation/export checks, graphics confinement, transcript golds, documentation disposition, module-documentation presence, treecheck, and the gated engineering record. Implementation is treated as replaceable; observable behavior and declared structure remain authoritative.

Publication hygiene was also completed. Historical debugging artifacts `out` and `output`, which exposed local-machine paths, were removed from all reachable history. A current-tree and reachable-history scan found no secrets, credentials, personal paths, or email addresses in tracked content. Commit-author email metadata was intentionally retained as provenance. No anomalously large reachable blobs remain.

Final validation was completed in the authoritative Linux environment: `make validate` passed, and the rewritten `master` history and tags were synchronized successfully to the private Odin remote. The alpha-release-prep session is therefore COMPLETE. No repository-preparation work remains in this session; GitHub publication and optional release-binary packaging are separate publication actions.
