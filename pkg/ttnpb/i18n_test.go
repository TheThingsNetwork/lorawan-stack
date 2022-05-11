// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import (
	"strings"
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/i18n"
)

func TestEnumMessages(t *testing.T) {
	testCases := []struct {
		desc  string
		names map[int32]string
	}{
		{desc: "GrantType", names: GrantType_name},
		{desc: "State", names: State_name},
		{desc: "ContactType", names: ContactType_name},
		{desc: "ContactMethod", names: ContactMethod_name},
		{desc: "MType", names: MType_name},
		{desc: "JoinRequestType", names: JoinRequestType_name},
		{desc: "RejoinRequestType", names: RejoinRequestType_name},
		{desc: "CFListType", names: CFListType_name},
		{desc: "MACCommandIdentifier", names: MACCommandIdentifier_name},
		{desc: "LocationSource", names: LocationSource_name},
		{desc: "PayloadFormatter", names: PayloadFormatter_name},
		{desc: "Right", names: Right_name},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			for _, s := range tc.names {
				if strings.ToLower(s) == s {
					continue // ignore "private" values.
				}
				if i18n.Get("enum:"+s) == nil {
					t.Errorf("message descriptor for %q is nil", s)
				}
			}
		})
	}
}
