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
	"context"
	"errors"
	"fmt"
	"io"
)

// Objects that implement this interface are used as middleware
// while sending and receiving messages. Typically this is used
// in support of symmetric capabilities such as CRC checksums
// or Reed Solomon parity bytes.
type Adapter interface {
	Wrap([]byte) ([]byte, error)
	Unwrap([]byte) ([]byte, error)
}

type FrameConfig struct {
	// Byte sequence that designates the start of a
	// message frame.
	FrameSyncMarker []byte

	// Size of fully encoded messages to be transmitted or
	// received. It is assumed that either a consant message
	// size will be used, or some sort of padding will be
	// applied by an included Adapter. This value does NOT
	// include the length of the sync marker.
	FrameSize int

	// Adapters apply basic encoding/decoding capabilities
	// to as messages are converted to and from frames.
	Adapters []Adapter
}

func (cfg *FrameConfig) Err() error {
	if len(cfg.FrameSyncMarker) == 0 {
		return errors.New("FrameSyncMarker must be provided")
	}

	if cfg.FrameSize <= 0 {
		return errors.New("FrameSize must be greater than 0")
	}

	return nil
}

func NewFrameSender(cfg FrameConfig, dst io.Writer) (*FrameSender, error) {
	if err := cfg.Err(); err != nil {
		return nil, err
	}

	fs := FrameSender{
		cfg: cfg,
		dst: dst,
	}
	return &fs, nil
}

type FrameSender struct {
	cfg FrameConfig
	dst io.Writer
}

func (s *FrameSender) Send(msg []byte) error {
	frm := msg
	var err error
	for _, ad := range s.cfg.Adapters {
		frm, err = ad.Wrap(frm)
		if err != nil {
			return err
		}
	}

	frmN := len(frm)
	if frmN > s.cfg.FrameSize {
		return errors.New("encoded frame exceeds maximum size")
	}

	frmWithASM := append(s.cfg.FrameSyncMarker, frm...)
	wantN := len(frmWithASM)

	n, err := s.dst.Write(frmWithASM)
	if err != nil {
		return err
	}

	// NOTE(bcwaldon): not sure what we should do in this case, so an error
	// seems most appropriate for now.
	if n != wantN {
		return errors.New("partial write")
	}

	return nil
}

func NewFrameReceiver(cfg FrameConfig, src io.Reader) (*FrameReceiver, error) {
	if err := cfg.Err(); err != nil {
		return nil, err
	}

	fr := FrameReceiver{
		cfg: cfg,
		src: src,
	}
	return &fr, nil
}

type FrameReceiver struct {
	cfg FrameConfig
	src io.Reader

	// Used to asynchronously communicate errors
	err error
}

// Forwards received frames to provided channel.
// This function is designed to be called within a goroutine.
//
// Receiving does not begin automatically when the FrameReceiver
// is created. A caller must use Receive to actually start reading
// from the source.
//
// If a nonrecoverable error is encountered, this function will
// exit after storing the error on the object. The error will
// be accessible through a subsequent call to Err().
//
// If the provided Context is cancelled or the provided channel
// is closed, the Receive method will exit.
func (r *FrameReceiver) Receive(ctx context.Context, ch chan<- []byte) {
	frameReader := NewFrameReader(r.src, r.cfg.FrameSyncMarker, r.cfg.FrameSize)

	wantN := len(r.cfg.FrameSyncMarker) + r.cfg.FrameSize

	for {
		// seek to next sync marker
		if err := frameReader.Seek(); err != nil && err != io.EOF {
			r.err = fmt.Errorf("read failure: %v", err)
			return
		}

		// now ready sync marker and full frame
		frm := make([]byte, wantN)
		n, err := frameReader.Read(frm)
		if err != nil {
			if err == io.EOF {
				return
			}

			//TODO(bcwaldon): determine if this behavior is acceptable
			r.err = fmt.Errorf("read failure: %v", err)
			return
		}

		// Before proceeding further, we also need to check
		// if the context was cancelled. This could have triggered
		// an empty read in an effort to gracefully shut down.
		select {
		case <-ctx.Done():
			r.err = ctx.Err()
			return
		default:
		}

		if n == 0 {
			//TODO(bcwaldon): determine if this behavior is acceptable
			r.err = errors.New("read failure: empty read operation")
			return
		} else if n != wantN {
			//TODO(bcwaldon): determine if this behavior is acceptable
			r.err = errors.New("read failure: partial read")
			return
		}

		// must strip leading sync marker
		msg := frm[len(r.cfg.FrameSyncMarker):]

		// Apply all adapters in reverse order
		for i := len(r.cfg.Adapters) - 1; i >= 0; i-- {
			msg, err = r.cfg.Adapters[i].Unwrap(msg)
			if err != nil {
				//TODO(bcwaldon): determine if this behavior is acceptable - probably not!
				r.err = fmt.Errorf("message decode failure: %v", err)
				return
			}
		}

		//TODO(bcwaldon): handle send to closed/full channel
		ch <- msg
	}

	return
}

func (r *FrameReceiver) Err() error {
	return r.err
}
