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
	"reflect"
	"testing"
)

func TestSpaceframeConfigFrameSize(t *testing.T) {
	cfg := SpaceframeConfig{
		Type:            SPACEFRAME_TYPE_CSP,
		PayloadDataSize: 1024,
	}
	want := 1026
	got := cfg.FrameSize()
	if got != want {
		t.Errorf("unexpected result: want=%d got=%d", want, got)
	}
}

func TestSpaceframeHeaderEncode(t *testing.T) {
	sh := SpaceframeHeader{
		Type:   SPACEFRAME_TYPE_CSP,
		Length: 78,
	}

	if err := sh.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []byte{0x00, 0x4e}
	got := sh.ToBytes()

	if !reflect.DeepEqual(want, got) {
		t.Errorf("unexpected result: want=%x got=%x", want, got)
	}
}

func TestSpaceframeHeaderDecode(t *testing.T) {
	hb := []byte{0x00, 0xb4}

	want := SpaceframeHeader{
		Type:   SPACEFRAME_TYPE_CSP,
		Length: 180,
	}

	got := SpaceframeHeader{}
	if err := got.FromBytes(hb); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("unexpected result: want=%#v got=%#v", want, got)
	}
}

func TestSpaceframeHeaderErrors(t *testing.T) {
	tests := []SpaceframeHeader{
		// Length is too large
		SpaceframeHeader{Type: SPACEFRAME_TYPE_CSP, Length: 1025},

		// Length is negative
		SpaceframeHeader{Type: SPACEFRAME_TYPE_CSP, Length: -1},

		// Type is invalid
		SpaceframeHeader{Type: SpaceframeType(2), Length: 180},
	}

	for i, tt := range tests {
		if err := tt.Err(); err == nil {
			t.Errorf("case %d: expected err, received nil", i)
		}
	}
}

func TestSpaceframeEnframe(t *testing.T) {
	tests := []struct {
		msg  []byte
		cfg  SpaceframeConfig
		want []byte
	}{
		// basic payload with type=CSP
		{
			msg:  []byte("XYZ"),
			cfg:  SpaceframeConfig{Type: SPACEFRAME_TYPE_CSP, PayloadDataSize: 3},
			want: []byte{0x0, 0x3, 0x58, 0x59, 0x5a},
		},

		// padding added as needed
		{
			msg:  []byte("XYZ"),
			cfg:  SpaceframeConfig{Type: SPACEFRAME_TYPE_CSP, PayloadDataSize: 6},
			want: []byte{0x0, 0x3, 0x58, 0x59, 0x5a, 0x0, 0x0, 0x0},
		},
	}

	for i, tt := range tests {
		got, err := Enframe(tt.msg, &tt.cfg)
		if err != nil {
			t.Errorf("case %d: unexpected error: %v", i, err)
			continue
		}

		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("case %d: unexpected result: want=%x got=%x", i, tt.want, got)
			continue
		}
	}
}

func TestSpaceframeEnframeErrors(t *testing.T) {
	tests := []struct {
		msg []byte
		cfg SpaceframeConfig
	}{
		// msg too large
		{
			msg: []byte{0x1, 0x2, 0x3, 0x4},
			cfg: SpaceframeConfig{Type: SPACEFRAME_TYPE_CSP, PayloadDataSize: 3},
		},

		// invalid config type
		{
			msg: []byte{0x0, 0x2, 0x2},
			cfg: SpaceframeConfig{Type: SpaceframeType(2), PayloadDataSize: 6},
		},
	}

	for i, tt := range tests {
		_, err := Enframe(tt.msg, &tt.cfg)
		if err == nil {
			t.Errorf("case %d: expected error, got nil", i)
		}

	}
}

func TestSpaceframeDeframe(t *testing.T) {
	tests := []struct {
		frm  []byte
		cfg  SpaceframeConfig
		want []byte
	}{
		// basic payload with type=CSP
		{
			frm:  []byte{0x0, 0x3, 0x58, 0x59, 0x5a},
			cfg:  SpaceframeConfig{Type: SPACEFRAME_TYPE_CSP, PayloadDataSize: 3},
			want: []byte{0x58, 0x59, 0x5a},
		},

		// padding removed as needed
		{
			frm:  []byte{0x0, 0x3, 0x58, 0x59, 0x5a, 0x0, 0x0, 0x0},
			cfg:  SpaceframeConfig{Type: SPACEFRAME_TYPE_CSP, PayloadDataSize: 6},
			want: []byte{0x58, 0x59, 0x5a},
		},
	}

	for i, tt := range tests {
		got, err := Deframe(tt.frm, &tt.cfg)
		if err != nil {
			t.Errorf("case %d: unexpected error: %v", i, err)
			continue
		}

		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("case %d: unexpected result: want=%x got=%x", i, tt.want, got)
			continue
		}
	}
}

func TestSpaceframeDeframeErrors(t *testing.T) {
	tests := []struct {
		frm []byte
		cfg SpaceframeConfig
	}{
		// frame too small
		{
			frm: []byte{0x0, 0x3, 0x58, 0x59, 0x5a, 0x0, 0x0, 0x0, 0x0},
			cfg: SpaceframeConfig{Type: SPACEFRAME_TYPE_CSP, PayloadDataSize: 8},
		},

		// frame too large
		{
			frm: []byte{0x0, 0x3, 0x58, 0x59, 0x5a, 0x0, 0x0, 0x0, 0x0},
			cfg: SpaceframeConfig{Type: SPACEFRAME_TYPE_CSP, PayloadDataSize: 6},
		},

		// header length overflow (only 2 of 3 expected bytes available, indicative of configuration issue)
		{
			frm: []byte{0x0, 0x3, 0x58, 0x59},
			cfg: SpaceframeConfig{Type: SPACEFRAME_TYPE_CSP, PayloadDataSize: 2},
		},
	}

	for i, tt := range tests {
		_, err := Deframe(tt.frm, &tt.cfg)
		if err == nil {
			t.Errorf("case %d: expected error, got nil", i)
		}

	}
}
