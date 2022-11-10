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
	CRC32_CHECKSUM_LENGTH_BYTES = 4
)

func applyCRC32(v []byte) []byte {
	cv := crc32.ChecksumIEEE(v)
	cb := make([]byte, CRC32_CHECKSUM_LENGTH_BYTES)
	binary.BigEndian.PutUint32(cb, cv)
	return append(v, cb...)
}

func verifyAndRemoveCRC32(v []byte) ([]byte, error) {
	vl := len(v)

	if vl <= CRC32_CHECKSUM_LENGTH_BYTES {
		return nil, errors.New("too few bytes for CRC32 validation")
	}

	// extract the trailer
	cb := v[vl-CRC32_CHECKSUM_LENGTH_BYTES:]
	v = v[:vl-CRC32_CHECKSUM_LENGTH_BYTES]

	gotCksum := binary.BigEndian.Uint32(cb)

	wantCksum := crc32.ChecksumIEEE(v)
	if gotCksum != wantCksum {
		return nil, errors.New("CRC32 checksum mismatch")
	}

	return v, nil
}
