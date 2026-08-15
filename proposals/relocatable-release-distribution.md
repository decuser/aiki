# Proposal: Relocatable Aiki Release Distribution

Status: ACTIVE

## Problem

Aiki currently works naturally from its source tree, where the executable and
`lib/` directory share the current working directory. That is a development
arrangement, not a user installation model. A user should be able to unpack an
Aiki release in a stable location, add that one directory to `PATH`, and run the
REPL or scripts from anywhere.

The runtime already embeds the grammar and prelude, but module-registry discovery
still depends on distribution roots expressed relative to the current working
directory. A lower-level loader has an executable-relative fallback, so module
path authority is currently split.

## Contract

The supported alpha installation model is:

1. unpack the release archive anywhere;
2. add the unpacked release directory to `PATH`;
3. invoke `aiki` from any working directory.

The release directory has this minimum shape:

```text
aiki-<version>-<os>-<arch>/
  aiki
  lib/
  LICENSE
  README.md
```

If a shipped `vendor/` tree exists, it is included beside `lib/`.

No installer, environment variable, fixed prefix, or source checkout is required.

## Authority

The executable location is the distribution root at runtime. Shipped module
roots are derived from that location. For development, only the conventional
`./lib` and `./vendor` directories are additional named-package roots. The
working directory itself is never recursively scanned, so launching installed
Aiki from a directory that happens to contain neighboring Aiki source trees
cannot discover their packages accidentally. Explicit path imports continue to
resolve local files relative to the importing program.

The Makefile owns construction of the release distribution. Generated release
artifacts live beside the source tree, never inside it: the unpacked user
distribution and its `.tar.gz` archive are siblings of the development tree and
are not tracked source.

## Cuts

### Cut 1 — executable-relative module roots

- derive shipped `lib/` and `vendor/` roots from the executable directory;
- keep the root policy separately testable from `os.Executable()`;
- preserve development-tree discovery through explicit `./lib` and `./vendor`
  roots plus the user library, without recursively scanning `.`.

Gate: focused module-root tests.

### Cut 2 — generated release archive

- add `make dist`;
- package the binary, shipped modules, README, and license under one versioned
  top-level directory;
- write both that unpacked distribution and its archive beside, never inside,
  the source tree;
- remove the old binary-only `make install` target because it produces an
  incomplete runtime installation under the new contract.

Gate: archive inspection.

### Cut 3 — relocatability proof and user documentation

- add `make distcheck`;
- unpack the generated archive into a temporary prefix;
- run from an unrelated working directory using only `PATH`;
- execute a script that imports and uses a shipped module;
- document the unpack-and-PATH installation model and distinguish installed Aiki
  from `./aiki` in the development tree.

Gate: `make distcheck`, then full `make validate` on the authoritative tree.

## Non-goals

This project does not add package-manager integration, `/usr/local` installation,
a launcher script, `AIKI_HOME`, automatic shell configuration, multi-version
management, cross-compilation, signing, or platform installers.

It does not change language semantics or module names.

## Acceptance

The project is complete when `make dist` leaves an unpacked user distribution
and matching release archive beside the source tree, and that archive can be
unpacked elsewhere so that, from another unrelated directory, `aiki` found
through `PATH` can run a script that loads a shipped module. The source tree must
not be needed at runtime.
