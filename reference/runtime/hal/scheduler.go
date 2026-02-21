package hal

import "aiki/reference/semantics/value"

// Scheduler defines the concurrency model.
// Implementations can use goroutines (GoScheduler) or cooperative green threads.
type Scheduler interface {
	Spawn(fn func()) value.Value
	MakeChan() value.Value
	Send(ch *value.Channel, val value.Value) value.Value
	Recv(ch *value.Channel) value.Value
}

// GoScheduler uses Go's goroutines for concurrency.
type GoScheduler struct{}

func NewGoScheduler() *GoScheduler {
	return &GoScheduler{}
}

func (s *GoScheduler) Spawn(fn func()) value.Value {
	go fn()
	return value.True
}

func (s *GoScheduler) MakeChan() value.Value {
	return value.NewChannel()
}

func (s *GoScheduler) Send(ch *value.Channel, val value.Value) value.Value {
	ch.C <- val
	return value.True
}

func (s *GoScheduler) Recv(ch *value.Channel) value.Value {
	return <-ch.C
}
