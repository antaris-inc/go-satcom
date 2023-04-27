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

	"github.com/sigurn/crc16"
)

const (
	CRC16_CHECKSUM_LENGTH_BYTES = 2
)

type CRC16AdapterConfig struct {
	Algorithm crc16.Params
}

func NewCRC16Adapter(cfg CRC16AdapterConfig) (*CRC16Adapter, error) {
	if cfg.Algorithm.Poly == 0 {
		return nil, errors.New("CRC Algorithm must be set")
	}
	ad := CRC16Adapter{
		Table: crc16.MakeTable(cfg.Algorithm),
	}
	return &ad, nil
}

// Supports append and strip/verify CRC16 checksums. Implements the
// satcom.Adapter interface.
type CRC16Adapter struct {
	*crc16.Table
}

func (a *CRC16Adapter) Wrap(v []byte) ([]byte, error) {
	return append(v, a.MakeChecksum(v)...), nil
}

func (a *CRC16Adapter) MakeChecksum(v []byte) []byte {
	cv := crc16.Checksum(v, a.Table)
	cb := make([]byte, CRC16_CHECKSUM_LENGTH_BYTES)
	binary.BigEndian.PutUint16(cb, cv)
	return cb
}

func (a *CRC16Adapter) Unwrap(v []byte) ([]byte, error) {
	vl := len(v)
	if vl <= CRC16_CHECKSUM_LENGTH_BYTES {
		return nil, errors.New("too few bytes for CRC validation")
	}

	// extract the trailer
	got := v[vl-CRC16_CHECKSUM_LENGTH_BYTES:]
	v = v[:vl-CRC16_CHECKSUM_LENGTH_BYTES]

	want := a.MakeChecksum(v)

	if got[0] != want[0] || got[1] != want[1] {
		return nil, errors.New("CRC checksum mismatch")
	}

	return v, nil
}
