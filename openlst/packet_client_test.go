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

package openlst

import (
	"reflect"
	"testing"
)

func TestClientPacketHeaderEncode(t *testing.T) {
	ph := ClientPacketHeader{
		Length:         10,
		HardwareID:     755,
		SequenceNumber: 12,
		Destination:    212,
		CommandNumber:  57,
	}

	if err := ph.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []byte{0x0a, 0xf3, 0x02, 0x0c, 0x00, 0xd4, 0x39}
	got := ph.ToBytes()

	if !reflect.DeepEqual(want, got) {
		t.Errorf("unexpected result: want=%x got=%x", want, got)
	}
}

func TestClientPacketHeaderDecode(t *testing.T) {
	hdr := []byte{0x0d, 0xff, 0x03, 0x04, 0x00, 0xfd, 0x38}

	want := ClientPacketHeader{
		Length:         13,
		HardwareID:     1023,
		SequenceNumber: 4,
		Destination:    253,
		CommandNumber:  56,
	}

	got := ClientPacketHeader{}
	if err := got.FromBytes(hdr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("unexpected result: want=%#v got=%#v", want, got)
	}
}

func TestNewClientPacketToBytes_SmallFrame(t *testing.T) {
	dat := []byte{0x0a, 0x0b, 0x0c, 0x0d}

	p := NewClientPacket(
		ClientPacketHeader{
			HardwareID:     1023,
			SequenceNumber: 1,
			Destination:    253,
			CommandNumber:  56,
		},
		dat,
	)
	if err := p.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []byte{0x0b, 0xff, 0x03, 0x01, 0x00, 0xfd, 0x38, 0x0a, 0x0b, 0x0c, 0x0d}
	got := p.ToBytes()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected encoded format: want=% x got=% x", want, got)
	}
}

func TestNewClientPacketToBytes_TooMuchData(t *testing.T) {
	dat := make([]byte, 1024)

	p := NewClientPacket(
		ClientPacketHeader{
			HardwareID:     1023,
			SequenceNumber: 1,
			Destination:    253,
			CommandNumber:  56,
		},
		dat,
	)
	if err := p.Err(); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestClientPacketFromBytes_SmallFrame(t *testing.T) {
	val := []byte{0x0a, 0xff, 0x03, 0x04, 0x00, 0xfd, 0x38, 0x01, 0x02, 0x03}

	p := ClientPacket{}
	if err := p.FromBytes(val); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := p.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantData := []byte{0x01, 0x02, 0x03}
	if !reflect.DeepEqual(wantData, p.Data) {
		t.Fatalf("unexpected payload: want=% x got=% x", wantData, p.Data)
	}
}

func TestClientPacketFromBytes_EmptyFrame(t *testing.T) {
	val := []byte{0x07, 0xff, 0x03, 0x04, 0x00, 0xfd, 0x38}

	p := ClientPacket{}
	if err := p.FromBytes(val); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := p.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(p.Data) != 0 {
		t.Fatalf("expected empty payload, got=% x", p.Data)
	}
}
