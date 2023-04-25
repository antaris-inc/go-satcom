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

package adapter

import (
	"errors"
	"fmt"

	csp "github.com/antaris-inc/go-satcom/csp/v1"
)

func NewCSPv1Adapter(packetHeader csp.PacketHeader, maxDataSize int) *cspAdapter {
	return &cspAdapter{
		PacketHeader: packetHeader,
		MaxDataSize:  maxDataSize,
	}
}

type cspAdapter struct {
	csp.PacketHeader
	MaxDataSize int
}

func (a *cspAdapter) Wrap(msg []byte) ([]byte, error) {
	if len(msg) > a.MaxDataSize {
		return nil, errors.New("message too large")
	}

	p := csp.Packet{
		PacketHeader: a.PacketHeader,
		Data:         msg,
	}

	buf := p.ToBytes()

	return buf, nil
}

func (a *cspAdapter) Unwrap(msg []byte) ([]byte, error) {
	if len(msg) > csp.MaxPacketLength(a.MaxDataSize) {
		return nil, errors.New("message too large")
	}

	var p csp.Packet
	if err := p.FromBytes(msg); err != nil {
		return nil, fmt.Errorf("CSP parsing failed: %v", err)
	}

	//NOTE(bcwaldon): header is assumed to be handled upstream
	// so it is not currently utilized here.

	return p.Data, nil
}
