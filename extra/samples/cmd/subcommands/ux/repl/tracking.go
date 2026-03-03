package repl

import "io"

// TrackingWriter tracks whether output ended with newline.
type TrackingWriter struct {
	Out              io.Writer
	EndedWithNewline bool
}

func (t *TrackingWriter) Write(p []byte) (n int, err error) {
	n, err = t.Out.Write(p)
	if n > 0 {
		t.EndedWithNewline = p[n-1] == '\n'
	}
	return
}
