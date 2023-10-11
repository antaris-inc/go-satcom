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
	"reflect"
	"testing"
)

func TestFrameReaderSingleFrame(t *testing.T) {
	syncMarker := []byte{0x01, 0x02, 0x03}

	buf := new(bytes.Buffer)
	buf.Write(syncMarker)
	buf.Write([]byte{0x04, 0x05, 0x06})

	rd := NewFrameReader(buf, syncMarker, 128)

	if err := rd.Seek(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := make([]byte, 6)
	n, err := rd.Read(got)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 6 {
		t.Fatalf("expected 6 bytes read, got %d", n)
	}

	want := append(syncMarker, []byte{0x04, 0x05, 0x06}...)
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected result: want=% x, got=% x", want, got)
	}
}

// Simulates a client reading a partial frame to decide how much
// more data is needed before calling read again.
func TestFrameReaderDynamicLengthFrame(t *testing.T) {
	syncMarker := []byte{0x01, 0x02, 0x03}

	buf := new(bytes.Buffer)
	buf.Write(syncMarker)
	buf.Write([]byte{0x04, 0x05, 0x06})
	buf.Write([]byte{0x07, 0x08, 0x09})
	buf.Write([]byte{0x10, 0x11, 0x12})

	rd := NewFrameReader(buf, syncMarker, 128)

	if err := rd.Seek(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := make([]byte, 6)
	want := append(syncMarker, []byte{0x04, 0x05, 0x06}...)

	n, err := rd.Read(got)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if n != 6 {
		t.Fatalf("expected 6 bytes read, got %d", n)
	} else if !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected result: want=% x, got=% x", want, got)
	}

	got = make([]byte, 3)
	want = []byte{0x07, 0x08, 0x09}

	n, err = rd.Read(got)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if n != 3 {
		t.Fatalf("expected 3 bytes read, got %d", n)
	} else if !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected result: want=% x, got=% x", want, got)
	}
}

func TestFrameReaderLargeSeek(t *testing.T) {
	syncMarker := []byte{0x01, 0x02, 0x03}

	buf := new(bytes.Buffer)

	// write a large amount of data to source, forcing many seeks ahead of real sync marker
	for i := 100; i < 161; i++ {
		buf.Write([]byte{uint8(i)})
	}

	buf.Write(syncMarker)
	buf.Write([]byte{0x04, 0x05, 0x06})

	rd := NewFrameReader(buf, syncMarker, 128)

	if err := rd.Seek(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := make([]byte, 6)
	want := append(syncMarker, []byte{0x04, 0x05, 0x06}...)

	n, err := rd.Read(got)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if n != 6 {
		t.Fatalf("expected 6 bytes read, got %d", n)
	} else if !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected result: want=% x, got=% x", want, got)
	}
}

// Ensure a partial sync marker at the end of the read buffer
// is treated properly.
func TestFrameReaderPartialSyncMarker(t *testing.T) {
	syncMarker := []byte{0x01, 0x02}

	buf := new(bytes.Buffer)

	buf.Write([]byte{0x04, 0x05, 0x06, 0x07})
	buf.Write([]byte{0x08, 0x09, 0x10, 0x01}) // first read through will stop here, seeing first byte of sync marker
	buf.Write([]byte{0x02, 0x11, 0x12, 0x13}) // expect sync to progress until it sees the full sync marker

	rd := NewFrameReader(buf, syncMarker, 128)

	if err := rd.Seek(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := make([]byte, 4)
	want := append(syncMarker, []byte{0x11, 0x12}...)

	n, err := rd.Read(got)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if n != 4 {
		t.Fatalf("expected 4 bytes read, got %d", n)
	} else if !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected result: want=% x, got=% x", want, got)
	}
}
