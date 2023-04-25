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

package satlab

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	SPACEFRAME_ASM_LENGTH_BYTES = 4

	SPACEFRAME_HEADER_LENGTH_BYTES = 2

	// field lengths (# bits)
	FLEN_RES    = 1
	FLEN_TYPE   = 4
	FLEN_LENGTH = 11
)

var (
	SPACEFRAME_ASM = []byte{0x1A, 0xCF, 0xFC, 0x1D}

	SPACEFRAME_TYPE_CSP = SpaceframeType(0)

	// Not yet supported/tested
	//SPACEFRAME_TYPE_IP  = SpaceframeType(1)
)

type SpaceframeConfig struct {
	Type            SpaceframeType
	PayloadDataSize int

	// If true, handle ASM prepend/strip. Typically
	// this is handled out of band, so the default
	// behavior is disabled.
	ASMEnabled bool

	// If true, append a CRC32c checksum to frames.
	// Checksum verification is also required
	// during deframe.
	CRCEnabled bool
}

func (cfg *SpaceframeConfig) FrameSize() int {
	n := SPACEFRAME_HEADER_LENGTH_BYTES + cfg.PayloadDataSize
	if cfg.ASMEnabled {
		n += SPACEFRAME_ASM_LENGTH_BYTES
	}
	if cfg.CRCEnabled {
		n += CRC_CHECKSUM_LENGTH_BYTES
	}
	return n
}

type SpaceframeType int

type SpaceframeHeader struct {
	Type   SpaceframeType
	Length int
}

func (h *SpaceframeHeader) Err() error {
	if h.Type != SPACEFRAME_TYPE_CSP {
		return errors.New("type invalid")
	}

	// datasheet is explicit about this limit
	if h.Length < 0 || h.Length > 1024 {
		return errors.New("length out of range (0-1024)")
	}

	return nil
}

func (h *SpaceframeHeader) ToBytes() []byte {
	var header uint16

	cursor := 0

	cursor += FLEN_RES
	header |= (uint16(0) << (16 - cursor))

	cursor += FLEN_TYPE
	header |= (uint16(h.Type) << (16 - cursor))

	cursor += FLEN_LENGTH
	header |= (uint16(h.Length) << (16 - cursor))

	bs := make([]byte, SPACEFRAME_HEADER_LENGTH_BYTES)
	binary.BigEndian.PutUint16(bs, header)

	return bs
}

func (h *SpaceframeHeader) FromBytes(bs []byte) error {
	if len(bs) != SPACEFRAME_HEADER_LENGTH_BYTES {
		return errors.New("Spaceframe header length unexpected")
	}

	hdr := binary.BigEndian.Uint16(bs)

	var offset int

	val := hdr << offset
	_ = int(val >> (16 - FLEN_RES))
	offset += FLEN_RES

	val = hdr << offset
	h.Type = SpaceframeType(val >> (16 - FLEN_TYPE))
	offset += FLEN_TYPE

	val = hdr << offset
	h.Length = int(val >> (16 - FLEN_LENGTH))
	offset += FLEN_LENGTH

	return nil
}

func Enframe(msg []byte, cfg *SpaceframeConfig) ([]byte, error) {
	msgLen := len(msg)
	if msgLen > cfg.PayloadDataSize {
		return nil, errors.New("message too large")
	}

	var hdr SpaceframeHeader
	hdr.Type = cfg.Type
	hdr.Length = msgLen

	if err := hdr.Err(); err != nil {
		return nil, fmt.Errorf("Spaceframe header: %v", err)
	}

	// start frame with encoded header
	frm := hdr.ToBytes()

	// create padded message and append to frame
	pmsg := make([]byte, cfg.PayloadDataSize)
	copy(pmsg, msg)
	frm = append(frm, pmsg...)

	if cfg.CRCEnabled {
		frm = applyCRC(frm)
	}

	if cfg.ASMEnabled {
		// prepend ASM
		frm = append(SPACEFRAME_ASM, frm...)
	}

	return frm, nil
}

func Deframe(frm []byte, cfg *SpaceframeConfig) ([]byte, error) {
	if len(frm) != cfg.FrameSize() {
		return nil, errors.New("Spaceframe length unexpected")
	}

	if cfg.ASMEnabled {
		if bytes.Compare(frm[:SPACEFRAME_ASM_LENGTH_BYTES], SPACEFRAME_ASM) != 0 {
			return nil, errors.New("Spaceframe ASM missing or invalid")
		}

		// strip ASM
		frm = frm[SPACEFRAME_ASM_LENGTH_BYTES:]
	}

	var err error
	if cfg.CRCEnabled {
		frm, err = verifyAndRemoveCRC(frm)
		if err != nil {
			return nil, err
		}
	}

	hb := frm[:SPACEFRAME_HEADER_LENGTH_BYTES]

	var hdr SpaceframeHeader
	if err = hdr.FromBytes(hb); err != nil {
		return nil, err
	}

	if err = hdr.Err(); err != nil {
		return nil, fmt.Errorf("Spaceframe header: %v", err)
	}

	if hdr.Type != cfg.Type {
		return nil, errors.New("Spaceframe header type unexpected")
	}

	msgStart := SPACEFRAME_HEADER_LENGTH_BYTES
	msgEnd := SPACEFRAME_HEADER_LENGTH_BYTES + hdr.Length

	if msgEnd > len(frm) {
		return nil, errors.New("Spaceframe length does not match value in header")
	}

	msg := frm[msgStart:msgEnd]

	return msg, nil
}
