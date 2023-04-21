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

func TestVehiclePacketHeaderEncode(t *testing.T) {
	ph := VehiclePacketHeader{
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

func TestVehiclePacketHeaderDecode(t *testing.T) {
	hdr := []byte{0x0d, 0xff, 0x03, 0x04, 0x00, 0xfd, 0x38}

	want := VehiclePacketHeader{
		Length:         13,
		HardwareID:     1023,
		SequenceNumber: 4,
		Destination:    253,
		CommandNumber:  56,
	}

	got := VehiclePacketHeader{}
	if err := got.FromBytes(hdr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("unexpected result: want=%#v got=%#v", want, got)
	}
}
