// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package types_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestJoinNonce(t *testing.T) {
	for _, tc := range []struct {
		JoinNonce JoinNonce
		IsZero    bool
		String    string
	}{
		{
			JoinNonce{0x00, 0x00, 0x00},
			true,
			"000000",
		},
		{
			JoinNonce{0x20, 0x00, 0x2f},
			false,
			"20002F",
		},
		{
			JoinNonce{0x40, 0x00, 0xef},
			false,
			"4000EF",
		},
	} {
		a := assertions.New(t)

		a.So(tc.JoinNonce.IsZero(), should.Equal, tc.IsZero)
		a.So(tc.JoinNonce.String(), should.Equal, tc.String)
	}
}
