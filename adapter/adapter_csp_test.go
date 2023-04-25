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

package adapter

import (
	"reflect"
	"testing"

	csp "github.com/antaris-inc/go-satcom/csp/v1"
)

func TestCSPAdapterWrap_Success(t *testing.T) {
	hdr := csp.PacketHeader{
		Priority:        2,
		Destination:     11,
		DestinationPort: 40,
		Source:          10,
		SourcePort:      20,
	}
	mtu := 5

	ad := NewCSPv1Adapter(hdr, mtu)

	msg := []byte{0x11, 0x22, 0x33}
	got, err := ad.Wrap(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []byte{0x94, 0xba, 0x14, 0x00, 0x11, 0x22, 0x33}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected result: want=% x got=% x", want, got)
	}
}

func TestCSPAdapterWrap_TooLarge(t *testing.T) {
	hdr := csp.PacketHeader{
		Priority:        2,
		Destination:     11,
		DestinationPort: 40,
		Source:          10,
		SourcePort:      20,
	}
	mtu := 2

	ad := NewCSPv1Adapter(hdr, mtu)

	msg := []byte{0x11, 0x22, 0x33}
	_, err := ad.Wrap(msg)
	if err == nil {
		t.Fatalf("expected non-nil error")
	}
}

func TestCSPAdapterUnwrap_Success(t *testing.T) {
	hdr := csp.PacketHeader{
		// Values here are inconsequential on Unwrap
	}
	mtu := 5
	ad := NewCSPv1Adapter(hdr, mtu)

	tests := []struct {
		msg  []byte
		want []byte
	}{
		// no data
		{
			msg:  []byte{0x94, 0xba, 0x14, 0x00},
			want: []byte{},
		},

		// message less than MTU
		{
			msg:  []byte{0x94, 0xba, 0x14, 0x00, 0x11, 0x22, 0x33},
			want: []byte{0x11, 0x22, 0x33},
		},

		// message at MTU
		{
			msg:  []byte{0x94, 0xba, 0x14, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			want: []byte{0x11, 0x22, 0x33, 0x44, 0x55},
		},
	}

	for ti, tt := range tests {
		got, err := ad.Unwrap(tt.msg)
		if err != nil {
			t.Errorf("case %d: unexpected error: %v", ti, err)
		}

		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("unexpected result: want=% x got=% x", tt.want, got)
		}
	}
}

func TestCSPAdapterUnwrap_Failure(t *testing.T) {
	hdr := csp.PacketHeader{
		// Values here are inconsequential on Unwrap
	}
	mtu := 5
	ad := NewCSPv1Adapter(hdr, mtu)

	tests := []struct {
		msg []byte
	}{
		// message too long
		{
			msg: []byte{0x94, 0xba, 0x14, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66},
		},

		// message header truncated
		{
			msg: []byte{0x94, 0xba},
		},

		// message empty
		{
			msg: []byte{},
		},
	}

	for ti, tt := range tests {
		_, err := ad.Unwrap(tt.msg)
		if err == nil {
			t.Errorf("case %d: expected non-nil error", ti)
		}

	}
}
