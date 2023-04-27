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
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/antaris-inc/go-satcom/crc"
	"github.com/antaris-inc/go-satcom/satlab"
)

func TestFrameSender_Success(t *testing.T) {
	crc32Adapter, _ := crc.NewCRC32Adapter(crc.CRC32AdapterConfig{
		Algorithm: crc.CRC32c,
	})

	tests := []struct {
		FrameConfig
		msg  []byte
		want []byte
	}{
		// Send max length w/o adapters
		{
			FrameConfig: FrameConfig{
				FrameSyncMarker: []byte{0xFF},
				FrameSize:       3,
				Adapters:        nil,
			},
			msg:  []byte{0x11, 0x22, 0x33},
			want: []byte{0xFF, 0x11, 0x22, 0x33},
		},

		// Send max length w/ one adapter
		{
			FrameConfig: FrameConfig{
				FrameSyncMarker: []byte{0xFF},
				FrameSize:       6,
				Adapters: []Adapter{
					crc32Adapter,
				},
			},
			msg:  []byte{0x11, 0x22},
			want: []byte{0xFF, 0x11, 0x22, 0x1C, 0x80, 0xE0, 0x0D},
		},

		// Send max length w/ two adapters
		{
			FrameConfig: FrameConfig{
				FrameSyncMarker: []byte{0xFE, 0xFF},
				FrameSize:       10,
				Adapters: []Adapter{
					&satlab.SatlabSpaceframeAdapter{
						satlab.SpaceframeConfig{
							Type:            satlab.SPACEFRAME_TYPE_CSP,
							PayloadDataSize: 4,
						},
					},
					crc32Adapter,
				},
			},
			msg: []byte{0x11, 0x22},
			want: []byte{
				0xFE, 0xFF, // ASM
				0x00, 0x02, // Spaceframe header
				0x11, 0x22, // original message
				0x00, 0x00, // Spaceframe padding
				0xBD, 0x02, 0x11, 0x4E, // CRC checksum
			},
		},
	}

	for ti, tt := range tests {
		buf := bytes.NewBuffer(nil)
		fs, err := NewFrameSender(tt.FrameConfig, buf)
		if err != nil {
			t.Errorf("case %d: failed constructing FrameSender: %v", ti, err)
			continue
		}
		if err := fs.Send(tt.msg); err != nil {
			t.Errorf("case %d: unexpected error: %v", ti, err)
		}
		got := buf.Bytes()
		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("case %d: unexpected result: want=% x got=% x", ti, tt.want, got)
		}
	}
}

func TestFrameSender_Failure(t *testing.T) {
	crc32Adapter, _ := crc.NewCRC32Adapter(crc.CRC32AdapterConfig{
		Algorithm: crc.CRC32c,
	})

	tests := []struct {
		FrameConfig
		msg []byte
	}{
		// Send over length w/o adapters
		{
			FrameConfig: FrameConfig{
				FrameSyncMarker: []byte{0xFF},
				FrameSize:       2,
				Adapters:        nil,
			},
			msg: []byte{0x11, 0x22, 0x33},
		},

		// Send over length w/ one adapter
		{
			FrameConfig: FrameConfig{
				FrameSyncMarker: []byte{0xFF},
				FrameSize:       5,
				Adapters: []Adapter{
					crc32Adapter,
				},
			},
			msg: []byte{0x11, 0x22, 0x33},
		},
	}

	for ti, tt := range tests {
		buf := bytes.NewBuffer(nil)
		fs, err := NewFrameSender(tt.FrameConfig, buf)
		if err != nil {
			t.Errorf("case %d: failed constructing Socket: %v", ti, err)
			continue
		}
		if err := fs.Send(tt.msg); err == nil {
			t.Errorf("case %d: expected non-nil error", ti)
		}
	}
}

func TestFrameReceiver_Success(t *testing.T) {
	crc32Adapter, _ := crc.NewCRC32Adapter(crc.CRC32AdapterConfig{
		Algorithm: crc.CRC32c,
	})

	tests := []struct {
		FrameConfig
		input []byte
		want  [][]byte
	}{
		// Three frames without adapters
		{
			FrameConfig: FrameConfig{
				FrameSyncMarker: []byte{0xFF},
				FrameSize:       3,
				Adapters:        nil,
			},
			input: []byte{
				0xFF, 0x11, 0x22, 0x33,
				0xFF, 0x44, 0x55, 0x66,
				0xFF, 0x77, 0x88, 0x99,
			},
			want: [][]byte{
				[]byte{0x11, 0x22, 0x33},
				[]byte{0x44, 0x55, 0x66},
				[]byte{0x77, 0x88, 0x99},
			},
		},

		// Two frames without adapters embedded in garbage
		{
			FrameConfig: FrameConfig{
				FrameSyncMarker: []byte{0xFF},
				FrameSize:       3,
				Adapters:        nil,
			},
			input: []byte{
				0xAA, 0xBB, 0xCC,
				0xFF, 0x44, 0x55, 0x66,
				0xFF, 0x77, 0x88, 0x99,
				0xDD, 0xEE,
			},
			want: [][]byte{
				[]byte{0x44, 0x55, 0x66},
				[]byte{0x77, 0x88, 0x99},
			},
		},

		// Two frames with adapter
		{
			FrameConfig: FrameConfig{
				FrameSyncMarker: []byte{0xFF},
				FrameSize:       6,
				Adapters: []Adapter{
					crc32Adapter,
				},
			},
			input: []byte{
				0xFF, 0x11, 0x22, 0x1C, 0x80, 0xE0, 0x0D,
				0xFF, 0x33, 0x44, 0x03, 0x29, 0x47, 0x6b,
			},
			want: [][]byte{
				[]byte{0x11, 0x22},
				[]byte{0x33, 0x44},
			},
		},
	}

	for ti, tt := range tests {
		buf := bytes.NewBuffer(tt.input)
		fr, err := NewFrameReceiver(tt.FrameConfig, buf)
		if err != nil {
			t.Errorf("case %d: unexpected error: %v", ti, err)
			continue
		}

		ch := make(chan []byte)
		go fr.Receive(context.Background(), ch)

		// iterate through expected frames and ensure we get a matching
		// frame from the receive channel
		for i := range tt.want {
			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			select {
			case got := <-ch:
				if !reflect.DeepEqual(got, tt.want[i]) {
					t.Errorf("case %d: frame %d: incorrect content: want=% x got=% x", ti, i, tt.want[i], got)
				}
			case <-ctx.Done():
				t.Errorf("case %d: failed to read expected frame %d in time", ti, i)
			}
		}

		// confirm no more frames are available
		select {
		case got := <-ch:
			t.Errorf("case %d: received unexpected additional frame: % x", ti, got)
		default:
		}

		if err := fr.Err(); err != nil {
			t.Errorf("case %d: frame receiver reported error: %v", ti, err)
		}

	}
}
