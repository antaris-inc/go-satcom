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

	"github.com/antaris-inc/go-satcom/adapter"
)

type MessageConfig struct {
	// Byte sequence that designates the start of a
	// message frame.
	FrameSyncMarker []byte

	// Maximum size of fully encoded messages that may
	// be transmitted or received, including sync marker
	// and any modifications made by adapters.
	FrameMTU int

	// Message adapters used to encode messages into
	// frames for transmission. Also used to adapt
	// frames back to messages on receipt.
	Adapters []adapter.Adapter
}

func (cfg *MessageConfig) Err() error {
	if len(cfg.FrameSyncMarker) == 0 {
		return errors.New("FrameSyncMarker must be provided")
	}

	if cfg.FrameMTU <= len(cfg.FrameSyncMarker) {
		return errors.New("FrameMTU must be greater than FrameSyncMarker length")
	}

	return nil
}

func NewMessageSender(cfg MessageConfig, dst io.Writer) (*MessageSender, error) {
	if err := cfg.Err(); err != nil {
		return nil, err
	}

	ms := MessageSender{
		cfg: cfg,
		dst: dst,
	}
	return &ms, nil
}

type MessageSender struct {
	cfg MessageConfig
	dst io.Writer
}

func (s *MessageSender) Send(msg []byte) error {
	enc := msg
	var err error
	for _, ad := range s.cfg.Adapters {
		enc, err = ad.Wrap(enc)
		if err != nil {
			return err
		}
	}

	enc = append(s.cfg.FrameSyncMarker, enc...)

	encLen := len(enc)
	if encLen > s.cfg.FrameMTU {
		return errors.New("encoded message exceeds MTU")
	}

	n, err := s.dst.Write(enc)
	if err != nil {
		return err
	}

	// NOTE(bcwaldon): not sure what we should do in this case, so an error
	// seems most appropriate for now.
	if n != encLen {
		return errors.New("partial write")
	}

	return nil
}

func NewMessageReceiver(cfg MessageConfig, src io.Reader) (*MessageReceiver, error) {
	if err := cfg.Err(); err != nil {
		return nil, err
	}

	mr := MessageReceiver{
		cfg: cfg,
		src: src,
	}
	return &mr, nil
}

type MessageReceiver struct {
	cfg MessageConfig
	src io.Reader

	// Used to asynchronously communicate errors
	err error
}

// Sends received messages to provided channel. This function is
// designed to be called within a goroutine.
//
// Receiving does not begin automatically when the MessageReceiver
// is created. A caller must use Receive to actually start reading
// from the source.
//
// If a nonrecoverable error is encountered, this function will
// exit after storing the error on the object. The error will
// be accessible through a subsequent call to Err().
//
// If the provided Context is cancelled or the provided channel
// is closed, the Receive method will exit.
func (r *MessageReceiver) Receive(ctx context.Context, ch chan<- []byte) {
	frameReader := NewFrameReader(r.src, r.cfg.FrameSyncMarker, r.cfg.FrameMTU)

	for {
		if err := frameReader.Seek(); err != nil {
			r.err = fmt.Errorf("read failure: %v", err)
			return
		}

		msg := make([]byte, r.cfg.FrameMTU)
		n, err := frameReader.Read(msg)
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
		}

		// must strip leading sync marker and truncate buffer to actual
		// number of bytes read
		msg = msg[len(r.cfg.FrameSyncMarker):n]
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

func (r *MessageReceiver) Err() error {
	return r.err
}
