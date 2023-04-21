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
	SPACE_PACKET_PREAMBLE = []byte{0xAA, 0xAA, 0xAA, 0xAA}
	SPACE_PACKET_ASM      = []byte{0xD3, 0x91, 0xD3, 0x91}

	SPACE_PACKET_HEADER_LENGTH = 6
	SPACE_PACKET_FOOTER_LENGTH = 4
)

type SpacePacketHeader struct {
	// Field 1
	Length int

	// Field 2 (a.k.a "Flags")
	Port int

	// Field 3
	SequenceNumber int

	// Field 4
	Destination int

	// Field 5
	CommandNumber int
}

func (p *SpacePacketHeader) Err() error {
	if p.Length < 7 || p.Length > 251 {
		return errors.New("Length must be 7-251")
	}
	if p.Port != 0 && p.Port != 1 {
		return errors.New("Port must be 0 or 1")
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

func (p *SpacePacketHeader) ToBytes() []byte {
	bs := make([]byte, SPACE_PACKET_HEADER_LENGTH)

	bs[0] = byte(p.Length)
	bs[1] = byte(p.Port)
	binary.LittleEndian.PutUint16(bs[2:4], uint16(p.SequenceNumber))
	bs[4] = byte(p.Destination)
	bs[5] = byte(p.CommandNumber)

	return bs
}

func (p *SpacePacketHeader) FromBytes(bs []byte) error {
	if len(bs) != SPACE_PACKET_HEADER_LENGTH {
		return errors.New("unexpected header length")
	}

	p.Length = int(bs[0])
	p.Port = int(bs[1])
	p.SequenceNumber = int(binary.LittleEndian.Uint16(bs[2:4]))
	p.Destination = int(bs[4])
	p.CommandNumber = int(bs[5])

	return nil
}

type SpacePacketFooter struct {
	// Field 1
	HardwareID int

	// Field 2
	CRC8 []byte
}

func (p *SpacePacketFooter) Err() error {
	if p.HardwareID < 0 || p.HardwareID > 65535 {
		return errors.New("HardwareID must be 0-65535")
	}
	if p.CRC8 == nil {
		return errors.New("CRC8 must be set")
	}
	return nil
}

func (p *SpacePacketFooter) ToBytes() []byte {
	bs := make([]byte, SPACE_PACKET_FOOTER_LENGTH)

	binary.LittleEndian.PutUint16(bs[0:2], uint16(p.HardwareID))

	// little endian mapping
	bs[2] = p.CRC8[1]
	bs[3] = p.CRC8[0]

	return bs
}

func (p *SpacePacketFooter) FromBytes(bs []byte) error {
	if len(bs) != SPACE_PACKET_FOOTER_LENGTH {
		return errors.New("unexpected footer length")
	}

	p.HardwareID = int(binary.LittleEndian.Uint16(bs[0:2]))

	// reversing little endian mapping
	p.CRC8 = make([]byte, 2)
	p.CRC8[0] = bs[3]
	p.CRC8[1] = bs[2]

	return nil
}
