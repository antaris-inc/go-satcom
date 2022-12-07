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
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	HEADER_LENGTH_BYTES = 4

	// field lengths (# bits)
	FLEN_PRIO  = 2
	FLEN_ADDR  = 5
	FLEN_PORT  = 6
	FLEN_FLAGS = 8
)

type PacketHeader struct {
	// 2 bits, conventionally:
	// 0 (critical), 1 (high), 2 (norm), 3 (low)
	Priority int

	// 5 bits: 0-31
	Destination int
	Source      int

	// 6 bits: 0-63
	DestinationPort int
	SourcePort      int

	// To be implemented
	// Flags int
}

func (p *PacketHeader) Err() error {
	if p.Priority < 0 || p.Priority > 3 {
		return errors.New("PacketHeader.Priority must be 0-3")
	}

	if p.Destination < 0 || p.Destination > 31 {
		return errors.New("PacketHeader.Destination must be 0-31")
	}
	if p.Source < 0 || p.Source > 31 {
		return errors.New("PacketHeader.Source must be 0-31")
	}

	if p.DestinationPort < 0 || p.DestinationPort > 63 {
		return errors.New("PacketHeader.DestinationPort must be 0-63")
	}
	if p.SourcePort < 0 || p.SourcePort > 63 {
		return errors.New("PacketHeader.SourcePort must be 0-63")
	}

	return nil
}

func (p *PacketHeader) ToBytes() []byte {
	var header uint32

	cursor := 0

	cursor += FLEN_PRIO
	header |= (uint32(p.Priority) << (32 - cursor))

	cursor += FLEN_ADDR
	header |= (uint32(p.Source) << (32 - cursor))

	cursor += FLEN_ADDR
	header |= (uint32(p.Destination) << (32 - cursor))

	cursor += FLEN_PORT
	header |= (uint32(p.DestinationPort) << (32 - cursor))

	cursor += FLEN_PORT
	header |= (uint32(p.SourcePort) << (32 - cursor))

	cursor += FLEN_FLAGS
	header |= (uint32(0) << (32 - cursor))

	bs := make([]byte, HEADER_LENGTH_BYTES)
	binary.BigEndian.PutUint32(bs, header)

	return bs
}

func (p *PacketHeader) FromBytes(bs []byte) error {
	if len(bs) != HEADER_LENGTH_BYTES {
		return errors.New("unexpected header length")
	}

	hdr := binary.BigEndian.Uint32(bs)

	var offset int

	val := hdr << offset
	p.Priority = int(val >> (32 - FLEN_PRIO))
	offset += FLEN_PRIO

	val = hdr << offset
	p.Source = int(val >> (32 - FLEN_ADDR))
	offset += FLEN_ADDR

	val = hdr << offset
	p.Destination = int(val >> (32 - FLEN_ADDR))
	offset += FLEN_ADDR

	val = hdr << offset
	p.DestinationPort = int(val >> (32 - FLEN_PORT))
	offset += FLEN_PORT

	val = hdr << offset
	p.SourcePort = int(val >> (32 - FLEN_PORT))
	offset += FLEN_PORT

	// not implemented, so ignored
	val = hdr << offset
	_ = int(val >> (32 - FLEN_FLAGS))
	offset += FLEN_FLAGS

	return nil
}

type Packet struct {
	PacketHeader
	Data []byte
}

func (p *Packet) Err() error {
	return p.PacketHeader.Err()
}

func (p *Packet) ToBytes() []byte {
	buf := MakeBuffer(len(p.Data))
	copy(buf, p.PacketHeader.ToBytes())
	copy(buf[HEADER_LENGTH_BYTES:], p.Data)
	return buf
}

func (p *Packet) FromBytes(bs []byte) error {
	if len(bs) < HEADER_LENGTH_BYTES {
		return errors.New("insufficient data")
	}

	hbs, dbs := bs[0:HEADER_LENGTH_BYTES], bs[HEADER_LENGTH_BYTES:]

	var ph PacketHeader
	if err := ph.FromBytes(hbs); err != nil {
		return err
	}

	p.PacketHeader = ph
	p.Data = dbs

	return nil
}

func WritePacket(dst io.Writer, p *Packet) error {
	enc := p.ToBytes()
	encLen := len(enc)

	n, err := dst.Write(enc)
	if err != nil {
		return err
	}

	if n != encLen {
		return fmt.Errorf("CSP write failed: want %d bytes, got %d", encLen, n)
	}

	return nil
}

// Initializes a new byte slice appropriate for a full CSP packet.
func MakeBuffer(maxDataSize int) []byte {
	return make([]byte, HEADER_LENGTH_BYTES+maxDataSize)
}

// Reads a CSP packet from an io.Reader using the supplied buffer. The caller should
// initialize the buffer to the max expected packet size (see MakeBuffer).
func ReadPacket(src io.Reader, buf []byte) (*Packet, error) {
	n, err := src.Read(buf)
	if err != nil {
		return nil, err
	}
	buf = buf[:n]

	var p Packet
	if err := p.FromBytes(buf); err != nil {
		return nil, fmt.Errorf("CSP parsing failed: %v", err)
	}

	return &p, nil
}
