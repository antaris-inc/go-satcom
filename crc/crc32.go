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

package crc

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
)

const (
	CRC32_CHECKSUM_LENGTH_BYTES = 4
)

// Keeping this here for convenience, as it is not obvious to all
// that "CRC32c" maps to the Castagnoli algorithm.
var CRC32c uint32 = crc32.Castagnoli

type CRC32AdapterConfig struct {
	Algorithm uint32
}

func NewCRC32Adapter(cfg CRC32AdapterConfig) (*CRC32Adapter, error) {
	if cfg.Algorithm == 0 {
		return nil, errors.New("CRC Algorithm must be set")
	}
	ad := CRC32Adapter{
		Table: crc32.MakeTable(cfg.Algorithm),
	}
	return &ad, nil
}

// Supports append and strip/verify CRC32 checksums. Implements the
// satcom.Adapter interface.
type CRC32Adapter struct {
	*crc32.Table
}

func (a *CRC32Adapter) MessageSize(n int) (int, error) {
	size := n + CRC32_CHECKSUM_LENGTH_BYTES
	return size, nil
}

func (a *CRC32Adapter) Wrap(v []byte) ([]byte, error) {
	cv := crc32.Checksum(v, a.Table)
	cb := make([]byte, CRC32_CHECKSUM_LENGTH_BYTES)
	binary.BigEndian.PutUint32(cb, cv)
	return append(v, cb...), nil
}

func (a *CRC32Adapter) Unwrap(v []byte) ([]byte, error) {
	vl := len(v)

	if vl <= CRC32_CHECKSUM_LENGTH_BYTES {
		return nil, errors.New("too few bytes for CRC validation")
	}

	// extract the trailer
	cb := v[vl-CRC32_CHECKSUM_LENGTH_BYTES:]
	v = v[:vl-CRC32_CHECKSUM_LENGTH_BYTES]

	gotCksum := binary.BigEndian.Uint32(cb)

	wantCksum := crc32.Checksum(v, a.Table)
	if gotCksum != wantCksum {
		return nil, errors.New("CRC checksum mismatch")
	}

	return v, nil
}
