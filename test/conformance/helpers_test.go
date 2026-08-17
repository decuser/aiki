package conformance

import (
	"path/filepath"
	"runtime"
	"testing"
)

func distributionRoot(t *testing.T) string {
	t.Helper()
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate conformance test source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(here), "..", ".."))
}
