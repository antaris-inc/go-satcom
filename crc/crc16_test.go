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

package crc

import (
	"reflect"
	"testing"

	"github.com/sigurn/crc16"
)

func TestCRC16Adapter_Wrap(t *testing.T) {
	ad, err := NewCRC16Adapter(CRC16AdapterConfig{Algorithm: crc16.CRC16_MAXIM})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gotBytes, err := ad.Wrap([]byte{0x1, 0x2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantBytes := []byte{0x1, 0x2, 0xae, 0x7f}

	if !reflect.DeepEqual(wantBytes, gotBytes) {
		t.Errorf("unexpected result: want=% x, got=% x", wantBytes, gotBytes)
	}
}

func TestCRC16Adapter_VerifySuccess(t *testing.T) {
	ad, err := NewCRC16Adapter(CRC16AdapterConfig{Algorithm: crc16.CRC16_MAXIM})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arg := []byte{0x1, 0x2, 0x03, 0x5E, 0xEF}
	gotBytes, err := ad.Unwrap(arg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantBytes := []byte{0x1, 0x2, 0x3}

	if !reflect.DeepEqual(wantBytes, gotBytes) {
		t.Errorf("unexpected result: want=%#v, got=%#v", wantBytes, gotBytes)
	}
}

func TestCRC16Adapter_VerifyFailure(t *testing.T) {
	ad, err := NewCRC16Adapter(CRC16AdapterConfig{Algorithm: crc16.CRC16_MAXIM})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := [][]byte{
		// Not enough data
		[]byte{0x1, 0x2},

		// CRC valid, but first byte zero'd out
		[]byte{0x0, 0x2, 0x03, 0xf8, 0x9f, 0x52},
	}

	for i, tt := range tests {
		_, err := ad.Unwrap(tt)
		if err == nil {
			t.Errorf("case %d: expected non-nil error", i)
		}
	}
}
