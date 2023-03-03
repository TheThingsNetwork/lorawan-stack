// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package alcsyncv1

import (
	"encoding/binary"
	"time"
)

// AppTimeReq is the device request for time correction.
type AppTimeReq struct {
	DeviceTime  time.Time
	TokenReq    uint8
	AnsRequired bool
}

// AppTimeAns is the answer to the device request for time correction.
type AppTimeAns struct {
	TimeCorrection int32
	TokenAns       uint8
}

// MarshalBinary marshals the AppTimeAns into a byte slice.
func (ans *AppTimeAns) MarshalBinary() ([]byte, error) {
	// DeviceTime - bytes [0,3].
	// Param - byte 4 (bits: RFU [7,4]; TokenAns [3,0]).

	cPayload := make([]byte, 5)
	binary.LittleEndian.PutUint32(cPayload, uint32(ans.TimeCorrection))
	cPayload[4] = ans.TokenAns & 0x0F
	return cPayload, nil
}
