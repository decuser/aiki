package substrate

import (
	"testing"

	"aiki/engine/semantics/value"
)

func addressParts(t *testing.T, v value.Value) (string, int64) {
	t.Helper()
	list, ok := v.(*value.List)
	if !ok || list.Shape != "address" || len(list.Elements) != 2 {
		t.Fatalf("address = %T %v", v, v.Inspect())
	}
	host := list.Elements[0].(*value.String).Val
	port := list.Elements[1].(*value.Number).Val.Num().Int64()
	return host, port
}

func TestNetworkTCPEndpointUsesIO(t *testing.T) {
	rt := NewGoRuntime()
	defer rt.CloseAllResources()
	listenerV := rt.halNetListen([]value.Value{&value.String{Val: "127.0.0.1"}, value.NewNumber(0, 1)}, nil)
	listener, ok := listenerV.(*value.Listener)
	if !ok {
		t.Fatalf("listen = %s", listenerV.Inspect())
	}
	_, port := addressParts(t, rt.halNetLocal([]value.Value{listener}, nil))
	accepted := make(chan value.Value, 1)
	go func() { accepted <- rt.halNetAccept([]value.Value{listener}, nil) }()
	clientV := rt.halNetConnect([]value.Value{&value.String{Val: "127.0.0.1"}, value.NewNumber(port, 1)}, nil)
	client, ok := clientV.(*value.Endpoint)
	if !ok {
		t.Fatalf("connect = %s", clientV.Inspect())
	}
	serverV := <-accepted
	server, ok := serverV.(*value.Endpoint)
	if !ok {
		t.Fatalf("accept = %s", serverV.Inspect())
	}
	if got := rt.halIOWrite([]value.Value{client, &value.String{Val: "ping\n"}}, nil); got != value.TRUE {
		t.Fatalf("write = %s", got.Inspect())
	}
	got := rt.halIOReadLine([]value.Value{server}, nil)
	if s, ok := got.(*value.String); !ok || s.Val != "ping" {
		t.Fatalf("read_line = %s", got.Inspect())
	}
}

func TestNetworkUDPDatagramPreservesBoundary(t *testing.T) {
	rt := NewGoRuntime()
	defer rt.CloseAllResources()
	recvV := rt.halNetUDPBind([]value.Value{&value.String{Val: "127.0.0.1"}, value.NewNumber(0, 1)}, nil)
	recv := recvV.(*value.Datagram)
	_, port := addressParts(t, rt.halNetLocal([]value.Value{recv}, nil))
	sendV := rt.halNetUDPBind([]value.Value{&value.String{Val: "127.0.0.1"}, value.NewNumber(0, 1)}, nil)
	send := sendV.(*value.Datagram)
	if got := rt.halNetUDPSend([]value.Value{send, &value.String{Val: "127.0.0.1"}, value.NewNumber(port, 1), &value.Bytes{Val: []byte("hello")}}, nil); got != value.TRUE {
		t.Fatalf("send = %s", got.Inspect())
	}
	packet := rt.halNetUDPRecv([]value.Value{recv}, nil)
	list, ok := packet.(*value.List)
	if !ok || list.Shape != "datagram" || len(list.Elements) != 3 {
		t.Fatalf("recv = %s", packet.Inspect())
	}
	b := list.Elements[0].(*value.Bytes)
	if string(b.Val) != "hello" {
		t.Fatalf("payload = %q", string(b.Val))
	}
}
