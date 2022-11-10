package v2

import (
	"encoding/binary"
	"errors"
)

const (
	HEADER_LENGTH_BYTES = 6

	// field lengths (# bits)
	FLEN_PRIO  = 2
	FLEN_ADDR  = 14
	FLEN_PORT  = 6
	FLEN_FLAGS = 6
)

type PacketHeader struct {
	// 2 bits, conventionally:
	// 0 (critical), 1 (high), 2 (norm), 3 (low)
	Priority int

	// 14 bits: 0-16383
	Destination int
	Source      int

	// 6 bits: 0-63
	DestinationPort int
	SourcePort      int

	// not yet implemented
	// Flags int
}

func (p *PacketHeader) Err() error {
	if p.Priority < 0 || p.Priority > 3 {
		return errors.New("PacketHeader.Priority must be 0-3")
	}

	if p.Destination < 0 || p.Destination > 16383 {
		return errors.New("PacketHeader.Destination must be 0-16383")
	}
	if p.Source < 0 || p.Source > 16383 {
		return errors.New("PacketHeader.Source must be 0-16383")
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
	var header uint64

	// starting at 48-bit offset given we are
	// packing a 48-bit header into a uint64
	cursor := 64 - 48

	cursor += FLEN_PRIO
	header |= (uint64(p.Priority) << (64 - cursor))

	cursor += FLEN_ADDR
	header |= (uint64(p.Destination) << (64 - cursor))

	cursor += FLEN_ADDR
	header |= (uint64(p.Source) << (64 - cursor))

	cursor += FLEN_PORT
	header |= (uint64(p.DestinationPort) << (64 - cursor))

	cursor += FLEN_PORT
	header |= (uint64(p.SourcePort) << (64 - cursor))

	cursor += FLEN_FLAGS
	header |= (uint64(0) << (64 - cursor))

	// convert to byte slice and discard first 2 bytes (48 bit header)
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, header)
	bs = bs[2:8]

	return bs
}

func (p *PacketHeader) FromBytes(bs []byte) error {
	if len(bs) != HEADER_LENGTH_BYTES {
		return errors.New("unexpected header length")
	}

	// need to pad 6 bytes up to 8 for uint64
	padding := []byte{0x0, 0x0}
	bs = append(padding, bs...)

	hdr := binary.BigEndian.Uint64(bs)

	// offset begins after first 2 bytes of uint64
	var offset = 64 - 48

	val := hdr << offset
	p.Priority = int(val >> (64 - FLEN_PRIO))
	offset += FLEN_PRIO

	val = hdr << offset
	p.Destination = int(val >> (64 - FLEN_ADDR))
	offset += FLEN_ADDR

	val = hdr << offset
	p.Source = int(val >> (64 - FLEN_ADDR))
	offset += FLEN_ADDR

	val = hdr << offset
	p.DestinationPort = int(val >> (64 - FLEN_PORT))
	offset += FLEN_PORT

	val = hdr << offset
	p.SourcePort = int(val >> (64 - FLEN_PORT))
	offset += FLEN_PORT

	// not implemented yet, so ignored
	val = hdr << offset
	_ = int(val >> (64 - FLEN_FLAGS))
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
	bs := p.PacketHeader.ToBytes()
	bs = append(bs, p.Data...)
	return bs
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
