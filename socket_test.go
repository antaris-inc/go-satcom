//   Copyright 2023 Antaris, Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package satcom

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/antaris-inc/go-satcom/adapter"
	csp "github.com/antaris-inc/go-satcom/csp/v1"
	"github.com/antaris-inc/go-satcom/satlab"
)

func TestSocketSend_Success(t *testing.T) {
	tests := []struct {
		SocketConfig
		msg  []byte
		want []byte
	}{
		// Send max MTU w/o adapters
		{
			SocketConfig: SocketConfig{
				MessageMTU: 4,
				Adapters:   nil,
				SyncMarker: []byte{0xFF},
			},
			msg:  []byte{0x11, 0x22, 0x33},
			want: []byte{0xFF, 0x11, 0x22, 0x33},
		},

		// Send max MTU w/ one adapter
		{
			SocketConfig: SocketConfig{
				MessageMTU: 7,
				Adapters: []adapter.Adapter{
					adapter.NewCSPv1Adapter(csp.PacketHeader{}, 2),
				},
				SyncMarker: []byte{0xFF},
			},
			msg:  []byte{0x11, 0x22},
			want: []byte{0xFF, 0x00, 0x00, 0x00, 0x00, 0x11, 0x22},
		},

		// Send max MTU w/ two adapters
		{
			SocketConfig: SocketConfig{
				MessageMTU: 11, // msg + CSP header + spaceframe header + ASM
				Adapters: []adapter.Adapter{
					adapter.NewCSPv1Adapter(csp.PacketHeader{
						Priority:        1,
						Source:          14,
						Destination:     15,
						SourcePort:      16,
						DestinationPort: 17,
					}, 4),
					&adapter.SatlabSpaceframeAdapter{
						satlab.SpaceframeConfig{
							PayloadDataSize: 8, // msg + CSP header
						},
					},
				},
				SyncMarker: []byte{0xFF},
			},
			msg: []byte{0x11, 0x22},
			want: []byte{
				0xFF,       // ASM
				0x00, 0x06, // Spaceframe header
				0x5c, 0xf4, 0x50, 0x00, // CSP header
				0x11, 0x22, // original message
				0x00, 0x00, // Spaceframe padding
			},
		},
	}

	for ti, tt := range tests {
		buf := bytes.NewBuffer(nil)
		sock, err := NewSocket(tt.SocketConfig, nil, buf)
		if err != nil {
			t.Errorf("case %d: failed constructing Socket: %v", ti, err)
			continue
		}
		if err := sock.Send(tt.msg); err != nil {
			t.Errorf("case %d: unexpected error: %v", ti, err)
		}
		got := buf.Bytes()
		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("case %d: unexpected result: want=% x got=% x", ti, tt.want, got)
		}
	}
}

func TestSocketSend_Failure(t *testing.T) {
	tests := []struct {
		SocketConfig
		msg []byte
	}{
		// Send over MTU w/o adapters
		{
			SocketConfig: SocketConfig{
				MessageMTU: 3,
				Adapters:   nil,
				SyncMarker: []byte{0xFF},
			},
			msg: []byte{0x11, 0x22, 0x33},
		},

		// Send over MTU w/ one adapter
		{
			SocketConfig: SocketConfig{
				MessageMTU: 6,
				Adapters: []adapter.Adapter{
					adapter.NewCSPv1Adapter(csp.PacketHeader{}, 2),
				},
				SyncMarker: []byte{0xFF},
			},
			msg: []byte{0x11, 0x22, 0x33},
		},
	}

	for ti, tt := range tests {
		buf := bytes.NewBuffer(nil)
		sock, err := NewSocket(tt.SocketConfig, nil, buf)
		if err != nil {
			t.Errorf("case %d: failed constructing Socket: %v", ti, err)
			continue
		}
		if err := sock.Send(tt.msg); err == nil {
			t.Errorf("case %d: expected non-nil error", ti)
		}
	}
}

func TestSocketRecv_Success(t *testing.T) {
	tests := []struct {
		SocketConfig
		input []byte
		want  [][]byte
	}{
		// Three frames without adapters
		{
			SocketConfig: SocketConfig{
				MessageMTU: 4,
				SyncMarker: []byte{0xFF},
				Adapters:   []adapter.Adapter{},
			},
			input: []byte{
				0xFF, 0x11, 0x22, 0x33,
				0xFF, 0x44, 0x55, 0x66,
				0xFF, 0x77, 0x88, 0x99,
			},
			want: [][]byte{
				[]byte{0x11, 0x22, 0x33},
				[]byte{0x44, 0x55, 0x66},
				[]byte{0x77, 0x88, 0x99},
			},
		},

		// Two frames without adapters embedded in garbage
		{
			SocketConfig: SocketConfig{
				MessageMTU: 4,
				SyncMarker: []byte{0xFF},
				Adapters:   []adapter.Adapter{},
			},
			input: []byte{
				0xAA, 0xBB, 0xCC,
				0xFF, 0x44, 0x55, 0x66,
				0xFF, 0x77, 0x88, 0x99,
				0xDD, 0xEE,
			},
			want: [][]byte{
				[]byte{0x44, 0x55, 0x66},
				[]byte{0x77, 0x88, 0x99},
			},
		},

		// Two frames with adapter
		{
			SocketConfig: SocketConfig{
				MessageMTU: 8,
				SyncMarker: []byte{0xFF},
				Adapters: []adapter.Adapter{
					adapter.NewCSPv1Adapter(csp.PacketHeader{}, 3),
				},
			},
			input: []byte{
				0xFF, 0x00, 0x00, 0x00, 0x00, 0x11, 0x22, 0x33,
				0xFF, 0x00, 0x00, 0x00, 0x00, 0x44, 0x55, 0x66,
			},
			want: [][]byte{
				[]byte{0x11, 0x22, 0x33},
				[]byte{0x44, 0x55, 0x66},
			},
		},
	}

	for ti, tt := range tests {
		buf := bytes.NewBuffer(tt.input)
		sock, err := NewSocket(tt.SocketConfig, buf, nil)
		if err != nil {
			t.Errorf("case %d: unexpected error: %v", ti, err)
			continue
		}

		got := [][]byte{}
		for msg := range sock.Recv(context.Background()) {
			got = append(got, msg)
		}

		if len(got) != len(tt.want) {
			t.Errorf("case %d: received unexpected number of frames: got=%d want=%d", ti, len(got), len(tt.want))
		}

		for i, gotFrame := range got {
			wantFrame := tt.want[i]
			if !reflect.DeepEqual(gotFrame, wantFrame) {
				t.Errorf("case %d: frame %d: incorrect content: want=% x got=% x", ti, i, wantFrame, gotFrame)
			}
		}
	}
}

// Confirm basic loopback functionality mimicking a Satlab transceiver
func TestSocketLoopback_Satlab(t *testing.T) {
	buf := bytes.NewBuffer(nil)

	cfg := SocketConfig{
		MessageMTU: 227,
		SyncMarker: satlab.SPACEFRAME_ASM,
		Adapters: []adapter.Adapter{
			adapter.NewCSPv1Adapter(csp.PacketHeader{
				Priority:        1,
				Source:          14,
				Destination:     15,
				SourcePort:      16,
				DestinationPort: 17,
			}, 213),
			&adapter.SatlabSpaceframeAdapter{
				satlab.SpaceframeConfig{
					PayloadDataSize: 217,
					CRCEnabled:      true,
				},
			},
		},
	}

	sock, err := NewSocket(cfg, buf, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// write message and assert sanity

	msg := []byte("XXX")
	if err := sock.Send(msg); err != nil {
		t.Fatalf("send operation failed: %v", err)
	}

	wantWrittenLen := 227

	gotWrittenBytes := buf.Bytes()
	gotWrittenLen := len(gotWrittenBytes)
	if gotWrittenLen != wantWrittenLen {
		t.Fatalf("wrote incorrect number of bytes: want=%d got=%d", wantWrittenLen, gotWrittenLen)
	}

	// read back the same message and assert sanity

	got := <-sock.Recv(context.Background())
	gotReadLen := len(got)

	wantReadLen := 3
	if gotReadLen != wantReadLen {
		t.Errorf("read incorrect number of bytes: want=%d got=%d", wantReadLen, gotReadLen)
	}

	if !reflect.DeepEqual(msg, got) {
		t.Fatalf("read incorrect bytes: want=%x got=%x", msg, got)
	}
}
