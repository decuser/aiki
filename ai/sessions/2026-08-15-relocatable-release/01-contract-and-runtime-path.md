# 01 — Distribution contract and runtime path

Status: IMPLEMENTED; gate pending authoritative toolchain.

## Finding

The grammar and prelude are embedded, but the module registry constructed its
shipped `lib` and `vendor` roots as current-working-directory-relative paths.
That makes the normal source-tree layout accidentally participate in runtime
correctness. A lower-level module loader already contained an executable-relative
fallback, so responsibility for locating shipped modules was split.

## Decision

The directory containing the running `aiki` executable is the distribution root.
The public installation model therefore needs no environment variable or fixed
filesystem prefix.

`ModuleRoots(homeDir, executableDir, workingDir)` makes the path policy
deterministic and unit-testable. `DefaultModuleRoots` obtains both
`os.Executable()` and the working directory and delegates to that policy. Named
packages are discovered only beneath explicit roots:

```text
<executable>/lib
<executable>/vendor
<working-directory>/lib
<working-directory>/vendor
<home>/.aiki/lib
```

Duplicate roots are collapsed. The development roots preserve `./aiki` use from
the source tree and allow tests/projects to provide conventional local package
roots, but the working directory itself is never recursively scanned. Local
files outside those package roots remain available through explicit path
imports.

An early implementation retained `.` as a registry root. Real distribution
testing from `~/forge/dev` exposed why that is unsafe: recursive discovery found
packages in both `aiki/` and a neighboring `aiki-ancient/` tree and reported
duplicate packages. A second implementation removed working-directory roots
entirely; full validation then exposed the opposite regression: development
execution and tests that intentionally use a local `lib/` could no longer find
modules. The final policy keeps only the explicit `lib`/`vendor` development
roots and therefore satisfies both contracts.

## Evidence

A focused unit test pins the expected roots:

```text
<executable>/lib
<executable>/vendor
<working-directory>/lib
<working-directory>/vendor
<home>/.aiki/lib
```

Compilation cannot be run in the current environment because Go attempts to
download the unavailable Go 1.24 toolchain.
