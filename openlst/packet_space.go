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
	"fmt"

	"github.com/antaris-inc/go-satcom/crc"
)

var (
	SPACE_PACKET_PREAMBLE = []byte{0xAA, 0xAA, 0xAA, 0xAA}
	SPACE_PACKET_ASM      = []byte{0xD3, 0x91, 0xD3, 0x91}

	SPACE_PACKET_HEADER_LENGTH = 6
	SPACE_PACKET_FOOTER_LENGTH = 4
)

var crc16Adapter *crc.CRC16Adapter

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
	minLen := SPACE_PACKET_HEADER_LENGTH + SPACE_PACKET_FOOTER_LENGTH - 1
	if p.Length < minLen || p.Length > 251 {
		return fmt.Errorf("Length must be %d-251", minLen)
	}
	//TODO(bcwaldon): confirm is this is actually true, as we see
	// a "flags" value of 192 used in gr-openlst
	//if p.Port != 0 && p.Port != 1 {
	//	return errors.New("Port must be 0 or 1")
	//}
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
	CRC16 []byte
}

func (p *SpacePacketFooter) Err() error {
	if p.HardwareID < 0 || p.HardwareID > 65535 {
		return errors.New("HardwareID must be 0-65535")
	}
	if len(p.CRC16) != 2 {
		return errors.New("CRC16 set incorrectly")
	}
	return nil
}

func (p *SpacePacketFooter) ToBytes() []byte {
	bs := make([]byte, SPACE_PACKET_FOOTER_LENGTH)

	binary.LittleEndian.PutUint16(bs[0:2], uint16(p.HardwareID))

	// nil check required as ToBytes is used during checksum creation.
	if p.CRC16 != nil {
		// little endian mapping
		bs[2] = p.CRC16[1]
		bs[3] = p.CRC16[0]
	}

	return bs
}

func (p *SpacePacketFooter) FromBytes(bs []byte) error {
	if len(bs) != SPACE_PACKET_FOOTER_LENGTH {
		return errors.New("unexpected footer length")
	}

	p.HardwareID = int(binary.LittleEndian.Uint16(bs[0:2]))

	// reversing little endian mapping
	p.CRC16 = make([]byte, 2)
	p.CRC16[0] = bs[3]
	p.CRC16[1] = bs[2]

	return nil
}

type SpacePacket struct {
	SpacePacketHeader
	Data []byte
	SpacePacketFooter
}

// Validates packet content, returning non-nil error if any issues detected.
func (p *SpacePacket) Err() error {
	if err := p.SpacePacketHeader.Err(); err != nil {
		return err
	}
	if err := p.SpacePacketFooter.Err(); err != nil {
		return err
	}

	if p.SpacePacketHeader.Length != SPACE_PACKET_HEADER_LENGTH+len(p.Data)+SPACE_PACKET_FOOTER_LENGTH-1 {
		return errors.New("packet length mismatch")
	}

	if err := p.verifyCRC16(); err != nil {
		return err
	}

	return nil
}

func (p *SpacePacket) verifyCRC16() error {
	got := p.SpacePacketFooter.CRC16
	want := makeSpacePacketCRC16(p)

	if got[0] != want[0] || got[1] != want[1] {
		return errors.New("checksum mismatch")
	}

	return nil
}

// Encodes packet to byte slice, including header, data and footer.
func (p *SpacePacket) ToBytes() []byte {
	buf := make([]byte, p.SpacePacketHeader.Length+1)
	copy(buf, p.SpacePacketHeader.ToBytes())
	copy(buf[SPACE_PACKET_HEADER_LENGTH:], p.Data)
	copy(buf[SPACE_PACKET_HEADER_LENGTH+len(p.Data):], p.SpacePacketFooter.ToBytes())
	return buf
}

// Hydrates Packet from provided byte slice, returning non-nil if any
// issues are encountered.
func (p *SpacePacket) FromBytes(bs []byte) error {
	if len(bs) < SPACE_PACKET_HEADER_LENGTH {
		return errors.New("insufficient data")
	}

	var ph SpacePacketHeader
	if err := ph.FromBytes(bs[0:SPACE_PACKET_HEADER_LENGTH]); err != nil {
		return err
	}

	var pf SpacePacketFooter
	if err := pf.FromBytes(bs[len(bs)-SPACE_PACKET_FOOTER_LENGTH:]); err != nil {
		return err
	}

	p.SpacePacketHeader = ph
	p.Data = bs[SPACE_PACKET_HEADER_LENGTH : len(bs)-SPACE_PACKET_FOOTER_LENGTH]
	p.SpacePacketFooter = pf

	return nil
}

// Constructs a new SpacePacket using provided header and data inputs.
//
// The header length and footer CRC16 fields are both set automatically
// based on the data provided. Length does not include itself.
//
// The packet returned must be confirmed as valid by the client before
// further use.
func NewSpacePacket(hdr SpacePacketHeader, dat []byte, ftr SpacePacketFooter) *SpacePacket {
	p := SpacePacket{
		SpacePacketHeader: hdr,
		Data:              dat,
		SpacePacketFooter: ftr,
	}

	p.SpacePacketHeader.Length = SPACE_PACKET_HEADER_LENGTH + len(dat) + SPACE_PACKET_FOOTER_LENGTH - 1
	p.SpacePacketFooter.CRC16 = makeSpacePacketCRC16(&p)

	return &p
}

// Calculates the CRC in the manner of the CC1110, which is documented here:
// 	 https://www.ti.com/lit/an/swra111e/swra111e.pdf?ts=1724970872901
func makeSpacePacketCRC16(p *SpacePacket) []byte {
	bs := p.ToBytes()

	// all fields considered (not leading preamble & ASM)
	inp := bs[0 : len(bs)-2]

	var ck uint16 = 0xFFFF
	for _, b := range inp {
		for i := 0; i < 8; i++ {
			if (((ck & 0x8000) >> 8) ^ uint16(b&0x80)) > 0 {
				ck = (ck << 1) ^ 0x8005
			} else {
				ck = ck << 1
			}
			b = b << 1
		}
	}

	ck = ck & 0xFFFF

	ckB := make([]byte, 2)
	binary.BigEndian.PutUint16(ckB, ck)
	return ckB
}
