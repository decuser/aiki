package host

type Scheduler interface {
	Spawn(fn func())
}

type Host interface {
	Scheduler() Scheduler
	Logf(format string, args ...any)
}

