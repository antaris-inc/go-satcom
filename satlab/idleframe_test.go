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
	"encoding/base64"
	"reflect"
	"testing"
)

func TestNewIdleFrameZeros(t *testing.T) {
	cfg := SpaceframeConfig{
		PayloadDataSize: 4,
		CRCEnabled:      true,
	}

	got, err := NewIdleFrameZeros(&cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("frm = % x", got)
	t.Logf("b64(frm) = %v", base64.StdEncoding.EncodeToString(got))

	if len(got) != cfg.FrameSize() {
		t.Fatalf("unexpected frame length")
	}

	got, err = Deframe(got, &cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []byte{0x0, 0x0, 0x0, 0x0}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected result: want=% x got=% x", want, got)
	}
}

func TestNewIdleFrameRand(t *testing.T) {
	cfg := SpaceframeConfig{
		PayloadDataSize: 4,
		CRCEnabled:      true,
	}

	got, err := NewIdleFrameRand(&cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("frm = % x", got)
	t.Logf("b64(frm) = %v", base64.StdEncoding.EncodeToString(got))

	if len(got) != cfg.FrameSize() {
		t.Fatalf("unexpected frame length")
	}

	_, err = Deframe(got, &cfg)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
