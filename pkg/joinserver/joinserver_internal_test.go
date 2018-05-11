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

package joinserver

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

func KeyPointer(key types.AES128Key) *types.AES128Key { return keyPointer(key) }

func TestMICCheck(t *testing.T) {
	a := assertions.New(t)

	pld := ttnpb.NewPopulatedJoinRequest(test.Randy, false).GetRawPayload()[:19]

	k := types.NewPopulatedAES128Key(test.Randy)
	computed, err := crypto.ComputeJoinRequestMIC(*k, pld)
	if err != nil {
		panic(err)
	}
	a.So(checkMIC(*k, append(pld, computed[:]...)), should.BeNil)
	a.So(checkMIC(*k, append(append(pld[1:], pld[0]-1), computed[:]...)), should.NotBeNil)
}
