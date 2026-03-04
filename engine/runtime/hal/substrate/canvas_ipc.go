package substrate

import "encoding/json"

// CanvasIPCMsg is a single message sent to the canvas child.
// Messages are encoded as one JSON object per line.
type CanvasIPCMsg struct {
	Proto int    `json:"proto"`
	Op    string `json:"op"`
	Args  []int  `json:"args,omitempty"`
	RGBA  []int  `json:"rgba,omitempty"`
	Pen   int    `json:"pen,omitempty"`
}

func (m CanvasIPCMsg) encodeLine() ([]byte, error) {
	if m.Proto == 0 {
		m.Proto = 1
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	b = append(b, '\n')
	return b, nil
}
