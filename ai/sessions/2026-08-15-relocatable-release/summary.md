# Summary — Relocatable Aiki release

Aiki's development layout had become its implicit installation layout. The
language executable was normally run from the repository root, where `lib/` was
naturally visible. For users, the desired model is much simpler and more durable:
unpack Aiki somewhere, add that directory to `PATH`, and run it.

The key architectural decision is that the executable identifies its own
distribution. Grammar and prelude were already embedded; shipped Aiki modules
now resolve relative to the executable as well. Named-package discovery no longer recursively scans the process working
directory. Installed Aiki uses the executable-relative `lib/` and `vendor/`;
development execution may additionally use only the explicit `./lib` and
`./vendor` roots; user packages live under `~/.aiki/lib`; other local files use
explicit path imports. This avoids accidental neighboring-tree discovery while
preserving the source-tree development workflow, without `AIKI_HOME`, fixed
installation prefixes, launch wrappers, or installer state.

The Makefile turns that contract into a reproducible unpacked distribution and
release archive, both written beside the source tree rather than into it.
`distcheck` tests the archive rather than merely testing the source tree.
The proof deliberately runs from an unrelated directory and imports a shipped
module. Thus the installation story is itself executable evidence.
