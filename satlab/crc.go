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
	"encoding/binary"
	"errors"
	"hash/crc32"
)

const (
	CRC_CHECKSUM_LENGTH_BYTES = 4
)

var (
	crc32cTable = crc32.MakeTable(crc32.Castagnoli)
)

func applyCRC(v []byte) []byte {
	cv := crc32.Checksum(v, crc32cTable)
	cb := make([]byte, CRC_CHECKSUM_LENGTH_BYTES)
	binary.BigEndian.PutUint32(cb, cv)
	return append(v, cb...)
}

func verifyAndRemoveCRC(v []byte) ([]byte, error) {
	vl := len(v)

	if vl <= CRC_CHECKSUM_LENGTH_BYTES {
		return nil, errors.New("too few bytes for CRC validation")
	}

	// extract the trailer
	cb := v[vl-CRC_CHECKSUM_LENGTH_BYTES:]
	v = v[:vl-CRC_CHECKSUM_LENGTH_BYTES]

	gotCksum := binary.BigEndian.Uint32(cb)

	wantCksum := crc32.Checksum(v, crc32cTable)
	if gotCksum != wantCksum {
		return nil, errors.New("CRC checksum mismatch")
	}

	return v, nil
}
