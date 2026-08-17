package substrate

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aiki/engine/semantics/value"
)

func nativeBytes(values ...int64) *value.List {
	elems := make([]value.Value, len(values))
	for i, n := range values {
		elems[i] = value.NewNumber(n, 1)
	}
	return &value.List{Shape: "bytes", Elements: elems}
}

func TestBytesFromAikiAcceptsOpaqueAndNative(t *testing.T) {
	opaque := &value.Bytes{Val: []byte{65, 66, 67}}
	got, err := bytesFromAiki(opaque)
	if err != nil || !bytes.Equal(got, []byte("ABC")) {
		t.Fatalf("opaque bytes: got %v, err %v", got, err)
	}

	got, err = bytesFromAiki(nativeBytes(65, 66, 67))
	if err != nil || !bytes.Equal(got, []byte("ABC")) {
		t.Fatalf("native bytes: got %v, err %v", got, err)
	}
}

func TestBytesFromAikiRejectsMalformedNativeBytes(t *testing.T) {
	cases := []struct {
		name string
		v    value.Value
		want string
	}{
		{"wrong shape", &value.List{Shape: "point", Elements: []value.Value{value.NewNumber(1, 1)}}, "expected bytes"},
		{"non integer", &value.List{Shape: "bytes", Elements: []value.Value{value.NewNumber(1, 2)}}, "must be an integer"},
		{"negative", nativeBytes(-1), "out of range"},
		{"too large", nativeBytes(256), "out of range"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := bytesFromAiki(tc.v)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got err %v, want containing %q", err, tc.want)
			}
		})
	}
}

func TestFileWriteBytesAcceptsNativeBytes(t *testing.T) {
	rt := NewGoRuntime()
	path := filepath.Join(t.TempDir(), "native.bin")
	opened := halFileOpen([]value.Value{&value.String{Val: path}, &value.Symbol{Val: "write"}}, nil)
	file := opened.(*value.File)
	if got := halFileWriteBytes([]value.Value{file, nativeBytes(0, 65, 255)}, nil); got != value.TRUE {
		t.Fatalf("write returned %s", got.Inspect())
	}
	rt.halFileClose([]value.Value{file}, nil)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte{0, 65, 255}) {
		t.Fatalf("got %v", got)
	}
}

func TestIOWriteAcceptsNativeBytes(t *testing.T) {
	var out bytes.Buffer
	rt := NewGoRuntime()
	rt.SetIO(nil, &out)
	if got := rt.halIOWrite([]value.Value{&value.Symbol{Val: "stdout"}, nativeBytes(65, 66, 67)}, nil); got != value.TRUE {
		t.Fatalf("write returned %s", got.Inspect())
	}
	if out.String() != "ABC" {
		t.Fatalf("got %q", out.String())
	}
}
