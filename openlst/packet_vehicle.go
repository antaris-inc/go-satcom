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
	"encoding/binary"
	"errors"
)

var (
	VEHICLE_PACKET_ASM = []byte{0x22, 0x69}

	VEHICLE_PACKET_HEADER_LENGTH = 7
)

type VehiclePacketHeader struct {
	// Field 1
	Length int

	// Field 2
	HardwareID int

	// Field 3
	SequenceNumber int

	// Field 4
	Destination int

	// Field 5
	CommandNumber int
}

func (p *VehiclePacketHeader) Err() error {
	if p.Length < 7 || p.Length > 251 {
		return errors.New("Length must be 7-251")
	}
	if p.HardwareID < 0 || p.HardwareID > 65535 {
		return errors.New("HardwareID must be 0-65535")
	}
	if p.SequenceNumber < 0 || p.SequenceNumber > 65535 {
		return errors.New("SequenceNumber must be 0-65535")
	}
	if p.Destination < 0 || p.Destination > 255 {
		return errors.New("Destination must be 0-255")
	}
	if p.CommandNumber < 0 || p.CommandNumber > 255 {
		return errors.New("CommandNumber must be 0-255")
	}
	return nil
}

func (p *VehiclePacketHeader) ToBytes() []byte {
	bs := make([]byte, VEHICLE_PACKET_HEADER_LENGTH)

	bs[0] = byte(p.Length)
	binary.LittleEndian.PutUint16(bs[1:3], uint16(p.HardwareID))
	binary.LittleEndian.PutUint16(bs[3:5], uint16(p.SequenceNumber))
	bs[5] = byte(p.Destination)
	bs[6] = byte(p.CommandNumber)

	return bs
}

func (p *VehiclePacketHeader) FromBytes(bs []byte) error {
	if len(bs) != VEHICLE_PACKET_HEADER_LENGTH {
		return errors.New("unexpected header length")
	}

	p.Length = int(bs[0])
	p.HardwareID = int(binary.LittleEndian.Uint16(bs[1:3]))
	p.SequenceNumber = int(binary.LittleEndian.Uint16(bs[3:5]))
	p.Destination = int(bs[5])
	p.CommandNumber = int(bs[6])

	return nil
}
