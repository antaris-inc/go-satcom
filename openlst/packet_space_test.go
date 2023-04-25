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

package openlst

import (
	"reflect"
	"testing"
)

func TestSpacePacketHeaderEncode(t *testing.T) {
	ph := SpacePacketHeader{
		Length:         27,
		Port:           0,
		SequenceNumber: 1134,
		Destination:    23,
		CommandNumber:  132,
	}

	if err := ph.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []byte{0x1b, 0x00, 0x6e, 0x04, 0x17, 0x84}
	got := ph.ToBytes()

	if !reflect.DeepEqual(want, got) {
		t.Errorf("unexpected result: want=%x got=%x", want, got)
	}
}

func TestSpacePacketHeaderDecode(t *testing.T) {
	hdr := []byte{0x0d, 0x01, 0x04, 0x00, 0xfd, 0x38}

	want := SpacePacketHeader{
		Length:         13,
		Port:           1,
		SequenceNumber: 4,
		Destination:    253,
		CommandNumber:  56,
	}

	got := SpacePacketHeader{}
	if err := got.FromBytes(hdr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("unexpected result: want=%#v got=%#v", want, got)
	}
}

func TestSpacePacketFooterEncode(t *testing.T) {
	pf := SpacePacketFooter{
		HardwareID: 2047,
		CRC16:      []byte{0x01, 0x02},
	}

	if err := pf.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []byte{0xff, 0x07, 0x02, 0x01}
	got := pf.ToBytes()

	if !reflect.DeepEqual(want, got) {
		t.Errorf("unexpected result: want=%x got=%x", want, got)
	}
}

func TestSpacePacketFooterDecode(t *testing.T) {
	ftr := []byte{0x0e, 0x01, 0x0b, 0x0a}

	want := SpacePacketFooter{
		HardwareID: 270,
		CRC16:      []byte{0x0a, 0x0b},
	}

	got := SpacePacketFooter{}
	if err := got.FromBytes(ftr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("unexpected result: want=%#v got=%#v", want, got)
	}
}

func TestNewSpacePacketToBytes_TooMuchData(t *testing.T) {
	p := NewSpacePacket(
		SpacePacketHeader{
			Port:           1,
			SequenceNumber: 4000,
			Destination:    253,
			CommandNumber:  56,
		},
		make([]byte, 1024),
		SpacePacketFooter{
			HardwareID: 12,
		},
	)
	if err := p.Err(); err == nil {
		t.Fatalf("expected non-nil error")
	}
}
