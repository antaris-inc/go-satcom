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
		sock, err := NewSocket(tt.SocketConfig, buf)
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
		sock, err := NewSocket(tt.SocketConfig, buf)
		if err != nil {
			t.Errorf("case %d: failed constructing Socket: %v", ti, err)
			continue
		}
		if err := sock.Send(tt.msg); err == nil {
			t.Errorf("case %d: expected non-nil error", ti)
		}
	}
}
