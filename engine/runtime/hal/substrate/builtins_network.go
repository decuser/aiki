package substrate

import (
	"errors"
	"net"
	"strconv"
	"sync"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

type ListenerResource struct {
	Listener net.Listener
	mu       sync.Mutex
	closed   bool
}

func (r *ListenerResource) close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil
	}
	r.closed = true
	return r.Listener.Close()
}

type DatagramResource struct {
	Conn   *net.UDPConn
	mu     sync.Mutex
	closed bool
}

func (r *DatagramResource) close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil
	}
	r.closed = true
	return r.Conn.Close()
}

func networkHostPort(args []value.Value, offset int, opname string) (string, int, *value.Fault) {
	if len(args) <= offset+1 {
		return "", 0, value.NewFault("%s: expected host and port", opname)
	}
	host, ok := args[offset].(*value.String)
	if !ok {
		return "", 0, value.NewFault("%s: host must be string, got %s", opname, args[offset].Type())
	}
	port, ok := args[offset+1].(*value.Number)
	if !ok || !port.IsInt() || port.Sign() < 0 || port.Compare(value.NewNumber(65535, 1)) > 0 {
		return "", 0, value.NewFault("%s: port must be an integer from 0 through 65535", opname)
	}
	return host.Val, int(port.Int64Value()), nil
}

func networkAddress(host string, port int) string {
	return net.JoinHostPort(host, strconv.Itoa(port))
}

func addressValue(addr net.Addr) value.Value {
	if addr == nil {
		return value.NewShapedError("network", "address unavailable")
	}
	host, portText, err := net.SplitHostPort(addr.String())
	if err != nil {
		return value.NewShapedError("network", "invalid host address %q: %v", addr.String(), err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		return value.NewShapedError("network", "invalid host port %q: %v", portText, err)
	}
	return &value.List{Shape: "address", Elements: []value.Value{
		&value.String{Val: host}, value.NewNumber(int64(port), 1),
	}}
}

func (g *GoRuntime) newNetworkEndpoint(label string, conn net.Conn) *value.Endpoint {
	ep := g.newEndpoint(label, conn, conn, conn)
	if resource, ok := g.endpointResource(ep); ok {
		resource.mu.Lock()
		resource.LocalAddr = conn.LocalAddr().String()
		resource.RemoteAddr = conn.RemoteAddr().String()
		resource.mu.Unlock()
	}
	return ep
}

func (g *GoRuntime) halNetConnect(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("net.connect: want 2 arguments, got %d", len(args))
	}
	host, port, fault := networkHostPort(args, 0, "net.connect")
	if fault != nil {
		return fault
	}
	conn, err := net.Dial("tcp", networkAddress(host, port))
	if err != nil {
		return value.NewShapedError("network", "connect: %v", err)
	}
	return g.newNetworkEndpoint("tcp", conn)
}

func (g *GoRuntime) halNetListen(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("net.listen: want 2 arguments, got %d", len(args))
	}
	host, port, fault := networkHostPort(args, 0, "net.listen")
	if fault != nil {
		return fault
	}
	listener, err := net.Listen("tcp", networkAddress(host, port))
	if err != nil {
		return value.NewShapedError("network", "listen: %v", err)
	}
	g.mu.Lock()
	g.nextListenerID++
	h := &value.Listener{ID: g.nextListenerID}
	g.listenerResources[h] = &ListenerResource{Listener: listener}
	g.mu.Unlock()
	return h
}

func (g *GoRuntime) listenerResource(h *value.Listener) (*ListenerResource, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	r, ok := g.listenerResources[h]
	return r, ok
}

func (g *GoRuntime) datagramResource(h *value.Datagram) (*DatagramResource, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	r, ok := g.datagramResources[h]
	return r, ok
}

func (g *GoRuntime) halNetAccept(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("net.accept: want 1 argument, got %d", len(args))
	}
	h, ok := args[0].(*value.Listener)
	if !ok {
		return value.NewFault("net.accept: expected listener, got %s", args[0].Type())
	}
	r, ok := g.listenerResource(h)
	if !ok {
		return value.NewShapedError("network", "listener does not belong to this runtime")
	}
	conn, err := r.Listener.Accept()
	if err != nil {
		return value.NewShapedError("network", "accept: %v", err)
	}
	return g.newNetworkEndpoint("tcp", conn)
}

func parseAddrString(s string) value.Value {
	host, portText, err := net.SplitHostPort(s)
	if err != nil {
		return value.NewShapedError("network", "invalid network address %q: %v", s, err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		return value.NewShapedError("network", "invalid network port %q: %v", portText, err)
	}
	return &value.List{Shape: "address", Elements: []value.Value{&value.String{Val: host}, value.NewNumber(int64(port), 1)}}
}

func (g *GoRuntime) halNetLocal(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("net.local: want 1 argument, got %d", len(args))
	}
	switch h := args[0].(type) {
	case *value.Listener:
		r, ok := g.listenerResource(h)
		if !ok {
			return value.NewShapedError("network", "listener does not belong to this runtime")
		}
		return addressValue(r.Listener.Addr())
	case *value.Datagram:
		r, ok := g.datagramResource(h)
		if !ok {
			return value.NewShapedError("network", "datagram socket does not belong to this runtime")
		}
		return addressValue(r.Conn.LocalAddr())
	case *value.Endpoint:
		r, ok := g.endpointResource(h)
		if !ok {
			return value.NewShapedError("network", "endpoint does not belong to this runtime")
		}
		r.mu.Lock()
		addr := r.LocalAddr
		r.mu.Unlock()
		if addr == "" {
			return value.NewShapedError("network", "endpoint has no network address")
		}
		return parseAddrString(addr)
	default:
		return value.NewFault("net.local: expected listener, datagram, or network endpoint, got %s", args[0].Type())
	}
}

func (g *GoRuntime) halNetRemote(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("net.remote: want 1 argument, got %d", len(args))
	}
	ep, ok := args[0].(*value.Endpoint)
	if !ok {
		return value.NewFault("net.remote: expected network endpoint, got %s", args[0].Type())
	}
	r, ok := g.endpointResource(ep)
	if !ok {
		return value.NewShapedError("network", "endpoint does not belong to this runtime")
	}
	r.mu.Lock()
	addr := r.RemoteAddr
	r.mu.Unlock()
	if addr == "" {
		return value.NewShapedError("network", "endpoint has no network address")
	}
	return parseAddrString(addr)
}

func (g *GoRuntime) halNetClose(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("net.close: want 1 argument, got %d", len(args))
	}
	var err error
	switch h := args[0].(type) {
	case *value.Listener:
		r, ok := g.listenerResource(h)
		if !ok {
			return value.NewShapedError("network", "listener does not belong to this runtime")
		}
		err = r.close()
	case *value.Datagram:
		r, ok := g.datagramResource(h)
		if !ok {
			return value.NewShapedError("network", "datagram socket does not belong to this runtime")
		}
		err = r.close()
	default:
		return value.NewFault("net.close: expected listener or datagram socket, got %s", args[0].Type())
	}
	if err != nil && !errors.Is(err, net.ErrClosed) {
		return value.NewShapedError("network", "close: %v", err)
	}
	return value.TRUE
}

func (g *GoRuntime) halNetUDPBind(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("net.udp_bind: want 2 arguments, got %d", len(args))
	}
	host, port, fault := networkHostPort(args, 0, "net.udp_bind")
	if fault != nil {
		return fault
	}
	addr, err := net.ResolveUDPAddr("udp", networkAddress(host, port))
	if err != nil {
		return value.NewShapedError("network", "udp_bind: %v", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return value.NewShapedError("network", "udp_bind: %v", err)
	}
	g.mu.Lock()
	g.nextDatagramID++
	h := &value.Datagram{ID: g.nextDatagramID}
	g.datagramResources[h] = &DatagramResource{Conn: conn}
	g.mu.Unlock()
	return h
}

func (g *GoRuntime) halNetUDPSend(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 4 {
		return value.NewFault("net.udp_send: want 4 arguments, got %d", len(args))
	}
	h, ok := args[0].(*value.Datagram)
	if !ok {
		return value.NewFault("net.udp_send: expected datagram socket, got %s", args[0].Type())
	}
	r, ok := g.datagramResource(h)
	if !ok {
		return value.NewShapedError("network", "datagram socket does not belong to this runtime")
	}
	host, port, fault := networkHostPort(args, 1, "net.udp_send")
	if fault != nil {
		return fault
	}
	data, err := bytesFromAiki(args[3])
	if err != nil {
		return value.NewFault("net.udp_send: %v", err)
	}
	addr, err := net.ResolveUDPAddr("udp", networkAddress(host, port))
	if err != nil {
		return value.NewShapedError("network", "udp_send: %v", err)
	}
	written, err := r.Conn.WriteToUDP(data, addr)
	if err != nil {
		return value.NewShapedError("network", "udp_send: %v", err)
	}
	if written != len(data) {
		return value.NewShapedError("network", "udp_send: short write: wrote %d of %d bytes", written, len(data))
	}
	return value.TRUE
}

func (g *GoRuntime) halNetUDPRecv(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("net.udp_recv: want 1 argument, got %d", len(args))
	}
	h, ok := args[0].(*value.Datagram)
	if !ok {
		return value.NewFault("net.udp_recv: expected datagram socket, got %s", args[0].Type())
	}
	r, ok := g.datagramResource(h)
	if !ok {
		return value.NewShapedError("network", "datagram socket does not belong to this runtime")
	}
	buf := make([]byte, 65535)
	n, addr, err := r.Conn.ReadFromUDP(buf)
	if err != nil {
		return value.NewShapedError("network", "udp_recv: %v", err)
	}
	host, portText, err := net.SplitHostPort(addr.String())
	if err != nil {
		return value.NewShapedError("network", "udp_recv: %v", err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		return value.NewShapedError("network", "udp_recv: %v", err)
	}
	return &value.List{Shape: "datagram", Elements: []value.Value{&value.Bytes{Val: buf[:n]}, &value.String{Val: host}, value.NewNumber(int64(port), 1)}}
}

func (g *GoRuntime) CloseAllNetwork() {
	g.mu.RLock()
	listeners := make([]*ListenerResource, 0, len(g.listenerResources))
	for _, r := range g.listenerResources {
		listeners = append(listeners, r)
	}
	datagrams := make([]*DatagramResource, 0, len(g.datagramResources))
	for _, r := range g.datagramResources {
		datagrams = append(datagrams, r)
	}
	g.mu.RUnlock()
	for _, r := range listeners {
		_ = r.close()
	}
	for _, r := range datagrams {
		_ = r.close()
	}
}
