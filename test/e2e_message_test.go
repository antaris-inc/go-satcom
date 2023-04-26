package main

import (
	"context"
	"net"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/antaris-inc/go-satcom"
	"github.com/antaris-inc/go-satcom/adapter"
	csp "github.com/antaris-inc/go-satcom/csp/v1"
	"github.com/antaris-inc/go-satcom/satlab"
)

func dialOrSkip(t *testing.T, env string) (net.Conn, error) {
	uplinkAddr := os.Getenv(env)
	if uplinkAddr == "" {
		t.Skip("skipping e2e tests")
	}

	var d net.Dialer
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	return d.DialContext(ctx, "tcp", uplinkAddr)
}

func TestE2EMessage_Loopback(t *testing.T) {
	uplink, err := dialOrSkip(t, "TEST_E2E_UPLINK_ADDRESS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer uplink.Close()

	downlink, err := dialOrSkip(t, "TEST_E2E_DOWNLINK_ADDRESS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer downlink.Close()

	cfg := satcom.MessageConfig{
		FrameMTU:        227,
		FrameSyncMarker: satlab.SPACEFRAME_ASM,
		Adapters: []adapter.Adapter{
			adapter.NewCSPv1Adapter(csp.PacketHeader{
				Priority:        1,
				Source:          14,
				SourcePort:      63,
				Destination:     18,
				DestinationPort: 7,
			}, 213),
			&adapter.SatlabSpaceframeAdapter{
				satlab.SpaceframeConfig{
					PayloadDataSize: 217,
					CRCEnabled:      true,
				},
			},
		},
	}

	ms, err := satcom.NewMessageSender(cfg, uplink)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mr, err := satcom.NewMessageReceiver(cfg, downlink)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	ch := make(chan []byte)
	go mr.Receive(ctx, ch)

	// write message

	msg := []byte("HELLO WORLD")
	if err := ms.Send(msg); err != nil {
		t.Fatalf("send operation failed: %v", err)
	}

	// Then read back the same message and assert the message loops back.
	// We cannot rely on Receive stopping on its own due to a cancelled context
	// since the context is not carried through the underlying Read operation.

	var got []byte

	select {
	case got = <-ch:
	case <-ctx.Done():
		t.Fatalf("failed to read message in time: %v", ctx.Err())
	}

	want := []byte("HELLO WORLD")
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("read incorrect bytes: want=% x got=% x", want, got)
	}
}
