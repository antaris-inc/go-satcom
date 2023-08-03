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
	// Given a payload, wrap it in the appropriate envelope
	Wrap([]byte) ([]byte, error)
	// Given a complete message, strip and verify expected envelope
	Unwrap([]byte) ([]byte, error)
	// Given a sample payload size, calculate the expected length of the wrapped message
	MessageSize(int) (int, error)
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

// Forward received frames to provided channel.
// This function is designed to be called within a goroutine.
// A caller must use Receive to actually start reading from
// the source.
//
// A call to Receive will continue until the upstream source
// is depleted, indicated by an io.EOF error. A user may also
// cancel the Receive operation prematurely by closing the
// provided Context.
//
// If a caller provides a non-nil error channel, any errors
// encountered within the frame processor will be sent to it.
// This channel is used synchronously, so a caller MUST read
// from it to unblock frame reception following an error.
func (r *FrameReceiver) Receive(ctx context.Context, msgC chan<- []byte, errC chan<- error) {
	frameReader := NewFrameReader(r.src, r.cfg.FrameSyncMarker, r.cfg.FrameSize)
	wantN := len(r.cfg.FrameSyncMarker) + r.cfg.FrameSize

	readFrame := func() ([]byte, error) { // Seek to next sync marker
		if err := frameReader.Seek(); err != nil {
			if err == io.EOF {
				return nil, err
			}
			return nil, fmt.Errorf("read failure: %v", err)
		}

		// now ready sync marker and full frame
		frm := make([]byte, wantN)
		n, err := frameReader.Read(frm)
		if err != nil {
			if err == io.EOF {
				return nil, err
			}

			return nil, fmt.Errorf("read failure: %v", err)
		}

		//TODO(bcwaldon): decide whether or not to check for canceled context again

		if n == 0 {
			return nil, errors.New("read failure: empty read operation")
		} else if n != wantN {
			return nil, errors.New("read failure: partial read")
		}

		// must strip leading sync marker
		msg := frm[len(r.cfg.FrameSyncMarker):]

		// Apply all adapters in reverse order
		for i := len(r.cfg.Adapters) - 1; i >= 0; i-- {
			msg, err = r.cfg.Adapters[i].Unwrap(msg)
			if err != nil {
				return nil, fmt.Errorf("decode failure: %v", err)

			}
		}

		return msg, nil
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg, err := readFrame()
		if err != nil {
			// signal to shut down, as the source is depleted
			if err == io.EOF {
				return
			}

			// Send an error back to the user if they provided a
			// channel for it. This allows the user to decide whether
			// or not to shut down the Receiver. The send operation
			// may block, which could have detrimental effects on
			// the upstream data source, but does give more control.
			if errC != nil {
				errC <- err
			}

			continue
		}

		select {
		// This send op may block, but it is up to the caller to
		// decide how to handle it.
		case msgC <- msg:
		case <-ctx.Done():
			return
		}
	}

	return
}
