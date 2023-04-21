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
		CRC8:       []byte{0x01, 0x02},
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
		CRC8:       []byte{0x0a, 0x0b},
	}

	got := SpacePacketFooter{}
	if err := got.FromBytes(ftr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("unexpected result: want=%#v got=%#v", want, got)
	}
}