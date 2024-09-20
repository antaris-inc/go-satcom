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
)

var (
	CLIENT_PACKET_ASM = []byte{0x22, 0x69}

	CLIENT_PACKET_HEADER_LENGTH = 7
)

type ClientPacketHeader struct {
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

func (p *ClientPacketHeader) Err() error {
	minLen := CLIENT_PACKET_HEADER_LENGTH - 1
	if p.Length < minLen || p.Length > 251 {
		return fmt.Errorf("Length must be %d-251", CLIENT_PACKET_HEADER_LENGTH)
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

func (p *ClientPacketHeader) ToBytes() []byte {
	bs := make([]byte, CLIENT_PACKET_HEADER_LENGTH)

	bs[0] = byte(p.Length)
	binary.LittleEndian.PutUint16(bs[1:3], uint16(p.HardwareID))
	binary.LittleEndian.PutUint16(bs[3:5], uint16(p.SequenceNumber))
	bs[5] = byte(p.Destination)
	bs[6] = byte(p.CommandNumber)

	return bs
}

func (p *ClientPacketHeader) FromBytes(bs []byte) error {
	if len(bs) != CLIENT_PACKET_HEADER_LENGTH {
		return errors.New("unexpected header length")
	}

	p.Length = int(bs[0])
	p.HardwareID = int(binary.LittleEndian.Uint16(bs[1:3]))
	p.SequenceNumber = int(binary.LittleEndian.Uint16(bs[3:5]))
	p.Destination = int(bs[5])
	p.CommandNumber = int(bs[6])

	return nil
}

type ClientPacket struct {
	ClientPacketHeader
	Data []byte
}

// Validates packet content, returning non-nil error if any issues detected.
func (p *ClientPacket) Err() error {
	if err := p.ClientPacketHeader.Err(); err != nil {
		return err
	}
	if p.ClientPacketHeader.Length != CLIENT_PACKET_HEADER_LENGTH+len(p.Data)-1 {
		return errors.New("packet length mismatch")
	}

	return nil
}

// Encodes packet to byte slice, including header and data.
func (p *ClientPacket) ToBytes() []byte {
	buf := make([]byte, p.ClientPacketHeader.Length+1)
	copy(buf, p.ClientPacketHeader.ToBytes())
	copy(buf[CLIENT_PACKET_HEADER_LENGTH:], p.Data)
	return buf
}

// Hydrates Packet from provided byte slice, returning non-nil if any
// issues are encountered.
func (p *ClientPacket) FromBytes(bs []byte) error {
	if len(bs) < CLIENT_PACKET_HEADER_LENGTH {
		return errors.New("insufficient data")
	}

	var ph ClientPacketHeader
	if err := ph.FromBytes(bs[0:CLIENT_PACKET_HEADER_LENGTH]); err != nil {
		return err
	}

	p.ClientPacketHeader = ph
	p.Data = bs[CLIENT_PACKET_HEADER_LENGTH:]

	return nil
}

// Constructs a new ClientPacket using provided header and data inputs.
//
// The header length field is automatically set based on the length of
// the provided data. Length does not include itself.
//
// The packet returned must be confirmed as valid by the client before
// further use.
func NewClientPacket(hdr ClientPacketHeader, dat []byte) *ClientPacket {
	p := ClientPacket{
		ClientPacketHeader: hdr,
		Data:               dat,
	}

	p.ClientPacketHeader.Length = CLIENT_PACKET_HEADER_LENGTH + len(dat) - 1

	return &p
}
