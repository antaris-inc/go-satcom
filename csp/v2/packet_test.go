package v2

import (
	"reflect"
	"testing"
)

func TestPacketHeaderEncode(t *testing.T) {
	ph := PacketHeader{
		Priority:        3,
		Destination:     2844,
		DestinationPort: 16,
		Source:          1728,
		SourcePort:      63,
	}

	if err := ph.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []byte{0xcb, 0x1c, 0x1b, 0x01, 0x0f, 0xc0}
	got := ph.ToBytes()

	if !reflect.DeepEqual(want, got) {
		t.Errorf("unexpected result: want=%x got=%x", want, got)
	}
}

func TestPacketHeaderDecode(t *testing.T) {
	hdr := []byte{0x80, 0x0c, 0x00, 0xac, 0x71, 0x40}

	want := PacketHeader{
		Priority:        2,
		Destination:     12,
		DestinationPort: 7,
		Source:          43,
		SourcePort:      5,
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
			Destination:     12,
			DestinationPort: 13,
			Source:          14,
			SourcePort:      15,
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
