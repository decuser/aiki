# Behavior Tests

Verify program outputs match expected results.

Smoke specimens use the `*_smoke.ai` suffix and have corresponding `.gold`
transcripts. Helper modules used by smoke specimens may also live here without
a gold file.

Run via: `aiki smoke` or `go test ./test/behavior/...`
