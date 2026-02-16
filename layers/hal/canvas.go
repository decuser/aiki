package hal

import (
	"sync"

	"aiki/lang/value"
)

var (
	openCanvases   []*value.Canvas
	openCanvasesMu sync.Mutex
)

func trackCanvas(c *value.Canvas) {
	openCanvasesMu.Lock()
	openCanvases = append(openCanvases, c)
	openCanvasesMu.Unlock()
}

func untrackCanvas(c *value.Canvas) {
	openCanvasesMu.Lock()
	for i, cvs := range openCanvases {
		if cvs == c {
			openCanvases = append(openCanvases[:i], openCanvases[i+1:]...)
			break
		}
	}
	openCanvasesMu.Unlock()
}

func CloseAllCanvases() {
	openCanvasesMu.Lock()
	for _, c := range openCanvases {
		select {
		case <-c.Done:
		default:
			close(c.Done)
		}
	}
	openCanvases = nil
	openCanvasesMu.Unlock()
}
