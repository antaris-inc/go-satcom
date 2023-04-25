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
	"errors"
	"io"

	"github.com/antaris-inc/go-satcom/adapter"
)

type Socket struct {
	Writer io.Writer

	MessageMTU int
	Adapters   []adapter.Adapter

	err error
}

func (c *Socket) Send(msg []byte) error {
	enc := msg
	var err error
	for _, ad := range c.Adapters {
		enc, err = ad.Wrap(enc)
		if err != nil {
			return err
		}
	}

	encLen := len(enc)
	if encLen > c.MessageMTU {
		return errors.New("encoded message exceeds MTU")
	}

	n, err := c.Writer.Write(enc)
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
