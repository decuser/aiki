package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSpawnPreservesDefiningFileForRelativeImport(t *testing.T) {
	withRepoRoot(t)

	dir := t.TempDir()
	depPath := filepath.Join(dir, "dep.ai")
	workerPath := filepath.Join(dir, "worker.ai")
	mainPath := filepath.Join(dir, "main.ai")

	if err := os.WriteFile(depPath, []byte(`package "spawn_relative_dep"
let answer = () { 42 }
export(:answer)
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(workerPath, []byte(`package "spawn_relative_worker"
let worker = (out) {
	let dep = import("./dep")
	send(out, dep.answer())
}
export(:worker)
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(mainPath, []byte(`let mod = import("./worker")
let out = channel()
spawn(mod.worker, out)
let got = recv(out)
if not equal(got, 42) { 1 / 0 }
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Run(mainPath); err != nil {
		t.Fatalf("spawned relative import: %v", err)
	}
}
