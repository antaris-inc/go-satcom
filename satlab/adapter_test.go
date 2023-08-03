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

import "testing"

func TestSpaceframeAdapter_MessageSize(t *testing.T) {
	ad := SpaceframeAdapter{
		SpaceframeConfig{
			Type:            SPACEFRAME_TYPE_CSP,
			PayloadDataSize: 12,
			WithASM:         true,
		},
	}

	wantSize := 18
	gotSize, err := ad.MessageSize(8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotSize != wantSize {
		t.Errorf("unexpected result: want=%v got=%v", wantSize, gotSize)
	}
}

func TestSpaceframeAdapter_MessageSize_TooLarge(t *testing.T) {
	ad := SpaceframeAdapter{
		SpaceframeConfig{
			Type:            SPACEFRAME_TYPE_CSP,
			PayloadDataSize: 12,
			WithASM:         true,
		},
	}

	_, err := ad.MessageSize(24)
	if err == nil {
		t.Fatalf("expected non-nil error")
	}
}
