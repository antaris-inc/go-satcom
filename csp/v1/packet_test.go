//   Copyright 2022 Antaris, Inc.
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

package v1

import (
	"reflect"
	"testing"
)

func TestPacketHeaderEncode(t *testing.T) {
	ph := PacketHeader{
		Priority:        2,
		Destination:     11,
		DestinationPort: 11,
		Source:          24,
		SourcePort:      11,
	}

	if err := ph.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []byte{0xb0, 0xb2, 0xcb, 0x00}
	got := ph.ToBytes()

	if !reflect.DeepEqual(want, got) {
		t.Errorf("unexpected result: want=%x got=%x", want, got)
	}
}

func TestPacketHeaderDecode(t *testing.T) {
	hdr := []byte{0xb0, 0xb2, 0xcb, 0x00}

	want := PacketHeader{
		Priority:        2,
		Destination:     11,
		DestinationPort: 11,
		Source:          24,
		SourcePort:      11,
	}

	got := PacketHeader{}
	if err := got.FromBytes(hdr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("unexpected result: want=%#v got=%#v", want, got)
	}
}

func TestPacketEncodeAndDecode(t *testing.T) {
	arg := Packet{
		PacketHeader: PacketHeader{
			Priority:        1,
			Destination:     2,
			DestinationPort: 3,
			Source:          4,
			SourcePort:      5,
		},
		Data: []byte("foobar"),
	}

	if err := arg.Err(); err != nil {
		t.Fatalf("unexpected error: err=%v", err)
	}

	gotBytes := arg.ToBytes()

	var got Packet
	if err := got.FromBytes(gotBytes); err != nil {
		t.Fatalf("unexpected error: err=%v", err)
	}

	want := arg
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected result: want=%v got=%v", want, got)
	}
}
