package repl

import "io"

type TrackingWriter struct {
	Out              io.Writer
	EndedWithNewline bool
}

func (w *TrackingWriter) Write(p []byte) (n int, err error) {
	n, err = w.Out.Write(p)
	if n > 0 {
		w.EndedWithNewline = p[n-1] == '\n'
	}
	return
}
