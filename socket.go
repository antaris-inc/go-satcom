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

type SocketConfig struct {
	SyncMarker []byte
	Adapters   []adapter.Adapter
	MessageMTU int
}

func NewSocket(cfg SocketConfig, downlink io.Reader, uplink io.Writer) (*Socket, error) {
	if len(cfg.SyncMarker) == 0 {
		return nil, errors.New("SyncMarker must be provided")
	}

	sock := Socket{
		cfg:    cfg,
		reader: downlink,
		writer: uplink,
	}
	return &sock, nil
}

type Socket struct {
	cfg    SocketConfig
	reader io.Reader
	writer io.Writer

	// Used to asynchronously communicate
	// errors during Recv
	recvErr error
}

func (c *Socket) Send(msg []byte) error {
	enc := msg
	var err error
	for _, ad := range c.cfg.Adapters {
		enc, err = ad.Wrap(enc)
		if err != nil {
			return err
		}
	}

	enc = append(c.cfg.SyncMarker, enc...)

	encLen := len(enc)
	if encLen > c.cfg.MessageMTU {
		return errors.New("encoded message exceeds MTU")
	}

	n, err := c.writer.Write(enc)
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

func (s *Socket) Recv(ctx context.Context) <-chan []byte {
	reader := NewFrameReader(s.reader, s.cfg.SyncMarker, s.cfg.MessageMTU)

	ch := make(chan []byte)
	go func() {
		defer func() {
			close(ch)
		}()

		for {
			if err := reader.Seek(); err != nil {
				s.recvErr = fmt.Errorf("read failure: %v", err)
				return
			}

			msg := make([]byte, s.cfg.MessageMTU)
			n, err := reader.Read(msg)

			// check if context is closed and shutdown if so
			select {
			case <-ctx.Done():
				s.recvErr = ctx.Err()
				return
			default:
			}

			if err != nil {
				if err == io.EOF {
					return
				}

				//TODO(bcwaldon): determine if this behavior is acceptable
				s.recvErr = fmt.Errorf("read failure: %v", err)
				return
			}

			if n == 0 {
				//TODO(bcwaldon): determine if this behavior is acceptable
				s.recvErr = errors.New("read failure: empty read operation")
				return
			}

			// must strip leading sync marker and truncate buffer to actual
			// number of bytes read
			msg = msg[len(s.cfg.SyncMarker):n]
			for i := len(s.cfg.Adapters) - 1; i >= 0; i-- {
				msg, err = s.cfg.Adapters[i].Unwrap(msg)
				if err != nil {
					//TODO(bcwaldon): determine if this behavior is acceptable - probably not!
					s.recvErr = fmt.Errorf("message decode failure: %v", err)
					return
				}
			}

			ch <- msg
		}
	}()

	return ch
}

func (c *Socket) RecvError() error {
	return c.recvErr
}
