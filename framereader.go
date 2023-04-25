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

package satcom

import (
	"bytes"
	"io"
)

func NewFrameReader(src io.Reader, syncMarker []byte, maxFrameSize int) *frameReader {
	return &frameReader{
		source:     src,
		syncMarker: syncMarker,
		readBuffer: make([]byte, 2*maxFrameSize),
	}

}

// Reads frames from a source io.Reader. The start of a frame is identified by
// first seeking through the source reader for the occurence of a sync marker.
// The frame end is identified by a subsequent sync marker, reaching the max
// frame length, or by receiving EOF from the source.
type frameReader struct {
	source     io.Reader
	syncMarker []byte

	readBuffer []byte
	cursor     int
}

// Read through enough data from the source to completely fill the provided byte slice.
// This operation will block until enough data is available or an error occurs.
func (c *frameReader) Read(dst []byte) (int, error) {
	dstN := len(dst)

	// ensure internal buffer is full enough to facilitate the read request
	if err := c.fillReadBuffer(dstN); err != nil {
		return 0, err
	}

	copy(dst, c.readBuffer[:dstN])
	c.seek(dstN)

	return dstN, nil
}

// Continue reading from source until a sync marker is identified. This will block
// until the a sync marker is found or the underlying source is depleted.
func (c *frameReader) Seek() error {
	for {
		// fill read buffer with at least enough data to check for the sync marker
		if err := c.fillReadBuffer(len(c.syncMarker)); err != nil {
			return err
		}

		// now, check the entire readBuffer (which may be much larger than the sync marker)
		idx := bytes.Index(c.readBuffer[:c.cursor], c.syncMarker)

		// sync marker found
		if idx >= 0 {
			// discard irrelevant data up to start of sync marker if necessary
			if idx >= 1 {
				c.seek(idx)
			}
			return nil
		}

		// no sync marker identified, so we discard all irrelevant data and repeat
		c.seek(c.cursor - len(c.syncMarker) + 1)
	}
}

func (c *frameReader) fillReadBuffer(target int) error {
	for {
		// enough data has already been buffered
		if c.cursor >= target {
			return nil
		}

		// read some more data, even if it is not enough to top up
		// the buffer to the target level (will repeat)
		n, err := c.source.Read(c.readBuffer[c.cursor:])
		if n == 0 {
			if err != nil {
				return err
			}

			//NOTE(bcwaldon): not sure how to handle this case yet
			panic("n==0")
		}

		c.cursor = c.cursor + n
	}
}

// Discard N leading bytes from buffer.
func (c *frameReader) seek(n int) {
	copy(c.readBuffer, c.readBuffer[n:])
	c.cursor = c.cursor - n
}
