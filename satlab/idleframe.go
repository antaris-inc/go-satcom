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
package satlab

import (
	"crypto/rand"
	"errors"
	"fmt"
)

// Creates an SRS4 idleframe using an "all zeros" message. This is a valid Spaceframe, but the
// payload is a series of zero values. Any configured features, such as CRC32, are implemented
// as intended (i.e. they are not zero'd out).
func NewIdleFrameZeros(cfg *SpaceframeConfig) ([]byte, error) {
	msg := make([]byte, cfg.PayloadDataSize)
	return Enframe(msg, cfg)
}

// Creates an SRS4 idleframe using pseudorandom data. This is NOT a valid Spaceframe. The frame
// length is based on the Spaceframe config, but the entire frame content (following the ASM)
// is pseudorandom data. Configured features, such as CRC32, will not validate on the receiving end.
func NewIdleFrameRand(cfg *SpaceframeConfig) ([]byte, error) {
	size := cfg.FrameSize()
	frm := make([]byte, size)

	want := len(SPACEFRAME_ASM)
	n := copy(frm[0:want], SPACEFRAME_ASM)
	if n != want {
		return nil, errors.New("header copy failed")
	}

	want = size - len(SPACEFRAME_ASM)
	var err error
	n, err = rand.Read(frm[len(SPACEFRAME_ASM):])
	if err != nil {
		return nil, fmt.Errorf("rand read failed: %v", err)
	} else if n != want {
		return nil, errors.New("read incorrect number of random bytes")
	}

	return frm, nil
}
