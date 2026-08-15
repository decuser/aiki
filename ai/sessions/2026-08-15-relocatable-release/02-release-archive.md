# 02 — Release archive and relocatability proof

Status: IMPLEMENTED; gate pending authoritative toolchain.

## Distribution

`make dist` now creates both a versioned, platform-labelled unpacked
distribution and its `.tar.gz` archive beside the source tree. The archive has
one top-level directory containing:

- `aiki`;
- `lib/`;
- `LICENSE`;
- `README.md`;
- `vendor/` when such a shipped tree exists.

The executable is deliberately at the top level so installation is exactly
"unpack and add this directory to PATH." No distribution staging directory is
created inside the repository.

The former `make install`, which copied only the executable to `~/bin`, was
removed because it would create an installation unable to carry Aiki's shipped
module library under the new explicit distribution contract.

## Distribution gate

`make distcheck` builds the archive, unpacks it under a temporary prefix, creates
an unrelated working directory, puts only the unpacked distribution on `PATH`,
and runs this program:

```aiki
use("list")
println(sum([1, 2, 3]))
```

The gate also plants two duplicate decoy packages beneath the unrelated working
directory. They must be ignored by named-package discovery. The expected program
output is `6`. This proves executable relocation, shipped-module discovery, and
independence from arbitrary package trees beneath the process working directory.

