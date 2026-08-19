#!/usr/bin/env python3
"""Verify that frozen pre-native/FFI behavioral witnesses still match baseline blobs."""
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parents[2]
manifest = root / "ai/evidence/native-ffi-baseline-witnesses.tsv"
allow = set()
if "--allow-post-preservation" in sys.argv[1:]:
    allowfile = root / "ai/evidence/native-ffi-post-preservation-allowlist.txt"
    for line in allowfile.read_text().splitlines():
        line = line.strip()
        if line and not line.startswith("#"):
            allow.add(line)
bad = []
changed_allowed = []
for line in manifest.read_text().splitlines():
    if not line or line.startswith("#"):
        continue
    blob, rel = line.split("\t", 1)
    path = root / rel
    if not path.exists():
        bad.append((rel, "MISSING", blob))
        continue
    cur = subprocess.check_output(["git", "-C", str(root), "hash-object", rel], text=True).strip()
    if cur != blob:
        if rel in allow:
            changed_allowed.append(rel)
        else:
            bad.append((rel, cur, blob))

if bad:
    print(f"FAIL: {len(bad)} frozen witness file(s) changed")
    for rel, cur, want in bad:
        print(f"  {rel}: {cur} != {want}")
    sys.exit(1)
if allow:
    print(f"PASS: no unapproved frozen-witness changes; {len(changed_allowed)} approved post-preservation change(s)")
    for rel in sorted(changed_allowed):
        print(f"  approved: {rel}")
else:
    print("PASS: 395 frozen baseline witness files unchanged")
