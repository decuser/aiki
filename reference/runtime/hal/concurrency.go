package hal

import "aiki/reference/semantics/value"

// DefaultScheduler is the package-level scheduler used by concurrency HAL.
// Can be swapped for testing or alternative implementations (e.g., green threads).
var DefaultScheduler Scheduler = NewGoScheduler()

// SetScheduler allows swapping the concurrency implementation.
func SetScheduler(s Scheduler) {
	DefaultScheduler = s
}

// Register concurrency HAL.
func init() {
	HAL["channel"] = &value.Builtin{
		Name: "channel",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 0 {
				return value.NewError("channel: want 0 arguments, got %d", len(args))
			}
			return DefaultScheduler.MakeChan()
		},
	}

	HAL["send"] = &value.Builtin{
		Name: "send",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 2 {
				return value.NewError("send: want 2 arguments, got %d", len(args))
			}
			ch, ok := args[0].(*value.Channel)
			if !ok {
				return value.NewError("send: first argument must be channel")
			}
			return DefaultScheduler.Send(ch, args[1])
		},
	}

	HAL["recv"] = &value.Builtin{
		Name: "recv",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 1 {
				return value.NewError("recv: want 1 argument, got %d", len(args))
			}
			ch, ok := args[0].(*value.Channel)
			if !ok {
				return value.NewError("recv: argument must be channel")
			}
			return DefaultScheduler.Recv(ch)
		},
	}

}
