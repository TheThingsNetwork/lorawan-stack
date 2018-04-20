// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package ttnpb

import "github.com/TheThingsNetwork/ttn/pkg/types"

const validIDChars = "abcdefghijklmnopqrstuvwxyz1234567890"

func NewPopulatedID(r randyIdentifiers) string {
	b := make([]byte, 2+r.Intn(35))
	for i := 0; i < len(b); i++ {
		b[i] = validIDChars[r.Intn(len(validIDChars))]
	}
	for n := 0; n < len(b)/8; n++ {
		i := 1 + r.Intn(len(b)-2)
		if b[i-1] != '_' && b[i-1] != '-' && b[i+1] != '_' && b[i+1] != '-' {
			b[i] = "-_"[r.Intn(2)]
		}
	}
	return string(b)
}

func NewPopulatedEndDeviceIdentifiers(r randyIdentifiers, easy bool) *EndDeviceIdentifiers {
	out := &EndDeviceIdentifiers{}
	if r.Intn(10) == 0 {
		out.DeviceID = NewPopulatedID(r)
	}
	out.ApplicationIdentifiers = ApplicationIdentifiers{
		ApplicationID: NewPopulatedID(r),
	}
	out.DevEUI = types.NewPopulatedEUI64(r)
	out.JoinEUI = types.NewPopulatedEUI64(r)
	out.DevAddr = types.NewPopulatedDevAddr(r)
	return out
}
