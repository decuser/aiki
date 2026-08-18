package value

import "fmt"

// Listener is an opaque Aiki-visible handle to a runtime-owned network listener.
type Listener struct{ ID uint64 }

func (l *Listener) Type() Type { return ListenerType }
func (l *Listener) Inspect() string {
	if l == nil || l.ID == 0 {
		return "<listener>"
	}
	return fmt.Sprintf("<listener:%d>", l.ID)
}

// Datagram is an opaque Aiki-visible handle to a runtime-owned datagram socket.
type Datagram struct{ ID uint64 }

func (d *Datagram) Type() Type { return DatagramType }
func (d *Datagram) Inspect() string {
	if d == nil || d.ID == 0 {
		return "<datagram>"
	}
	return fmt.Sprintf("<datagram:%d>", d.ID)
}
