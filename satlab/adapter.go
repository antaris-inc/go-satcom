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

import "errors"

type SpaceframeAdapter struct {
	SpaceframeConfig
}

func (a *SpaceframeAdapter) MessageSize(n int) (int, error) {
	if n > a.SpaceframeConfig.PayloadDataSize {
		return 0, errors.New("payload length limit exceeded")
	}

	return a.SpaceframeConfig.FrameSize(), nil
}

func (a *SpaceframeAdapter) Wrap(msg []byte) ([]byte, error) {
	return Enframe(msg, &a.SpaceframeConfig)
}

func (a *SpaceframeAdapter) Unwrap(frm []byte) ([]byte, error) {
	return Deframe(frm, &a.SpaceframeConfig)
}
