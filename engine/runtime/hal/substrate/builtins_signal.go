package substrate

import (
	"os"
	ossignal "os/signal"
	"sync"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

type SignalResource struct {
	host     chan os.Signal
	events   *value.Channel
	done     chan struct{}
	stopOnce sync.Once
}

func (r *SignalResource) stop() {
	r.stopOnce.Do(func() {
		ossignal.Stop(r.host)
		close(r.done)
	})
}

func signalNames(args []value.Value) ([]string, *value.Fault) {
	if len(args) != 1 {
		return nil, value.NewFault("signal.watch: want signal list, got %d arguments", len(args))
	}
	list, ok := args[0].(*value.List)
	if !ok {
		return nil, value.NewFault("signal.watch: expected list of symbols, got %s", args[0].Type())
	}
	if len(list.Elements) == 0 {
		return nil, value.NewFault("signal.watch: want at least one signal")
	}
	names := make([]string, len(list.Elements))
	for i, elem := range list.Elements {
		sym, ok := elem.(*value.Symbol)
		if !ok {
			return nil, value.NewFault("signal.watch: signal %d must be symbol, got %s", i, elem.Type())
		}
		names[i] = sym.Val
	}
	return names, nil
}

func (g *GoRuntime) halSignalWatch(args []value.Value, ctx *hal.EvalContext) value.Value {
	names, fault := signalNames(args)
	if fault != nil {
		return fault
	}
	hostSignals := make([]os.Signal, 0, len(names))
	for _, name := range names {
		sig, ok := hostSignalForWatch(name)
		if !ok {
			return value.NewShapedError("unsupported", "signal.watch: unsupported signal :%s on this host", name)
		}
		hostSignals = append(hostSignals, sig)
	}

	events := value.NewEventChannel()
	r := &SignalResource{host: make(chan os.Signal, 1), events: events, done: make(chan struct{})}
	g.mu.Lock()
	g.signalResources[events] = r
	g.mu.Unlock()
	ossignal.Notify(r.host, hostSignals...)

	go func() {
		for {
			select {
			case sig := <-r.host:
				name, ok := portableSignalName(sig)
				if !ok {
					continue
				}
				select {
				case r.events.C <- &value.Symbol{Val: name}:
				case <-r.done:
					return
				default:
					// Signal notification is edge-like and may coalesce. Never
					// strand the host notifier behind an unread event channel.
				}
			case <-r.done:
				return
			}
		}
	}()
	return events
}

func (g *GoRuntime) halSignalStop(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("signal.stop: want 1 argument, got %d", len(args))
	}
	ch, ok := args[0].(*value.Channel)
	if !ok {
		return value.NewFault("signal.stop: expected channel, got %s", args[0].Type())
	}
	g.mu.RLock()
	r, ok := g.signalResources[ch]
	g.mu.RUnlock()
	if !ok {
		return value.NewShapedError("signal", "signal source does not belong to this runtime")
	}
	r.stop()
	return value.TRUE
}

func (g *GoRuntime) halSignalSend(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("signal.send: want 2 arguments, got %d", len(args))
	}
	proc, ok := args[0].(*value.Process)
	if !ok {
		return value.NewFault("signal.send: expected process, got %s", args[0].Type())
	}
	sym, ok := args[1].(*value.Symbol)
	if !ok {
		return value.NewFault("signal.send: expected signal symbol, got %s", args[1].Type())
	}
	r, ok := g.processResource(proc)
	if !ok {
		return value.NewShapedError("process", "process does not belong to this runtime")
	}
	select {
	case <-r.done:
		return value.NewShapedError("process", "process has already exited")
	default:
	}
	if r.Cmd.Process == nil {
		return value.NewShapedError("process", "process has not started")
	}
	if err := sendPortableSignal(r.Cmd.Process, sym.Val); err != nil {
		if isUnsupportedSignal(err) {
			return value.NewShapedError("unsupported", "signal.send: unsupported signal :%s on this host", sym.Val)
		}
		return value.NewShapedError("signal", "%s", err.Error())
	}
	return value.TRUE
}

func (g *GoRuntime) CloseAllSignals() {
	g.mu.RLock()
	resources := make([]*SignalResource, 0, len(g.signalResources))
	for _, r := range g.signalResources {
		resources = append(resources, r)
	}
	g.mu.RUnlock()
	for _, r := range resources {
		r.stop()
	}
}
