package hal

import (
	"fmt"
	"os"

	"aiki/semantics/host"

)

type hostImpl struct{}

func NewHost() host.Host {
	return hostImpl{}
}

type schedulerWrap struct{}

func (schedulerWrap) Spawn(fn func()) {
	// DefaultScheduler.Spawn returns value.Value in your current hal
	_ = DefaultScheduler.Spawn(fn)
}

func (hostImpl) Scheduler() host.Scheduler {
	return schedulerWrap{}
}

func (hostImpl) Logf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
}

