// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package joinserver

import (
	"math/rand"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/crypto"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func KeyPointer(key types.AES128Key) *types.AES128Key { return keyPointer(key) }

func TestMICCheck(t *testing.T) {
	a := assertions.New(t)

	pld := ttnpb.NewPopulatedJoinRequest(test.Randy, false).GetRawPayload()[:19]

	k := types.NewPopulatedAES128Key(rand.New(rand.NewSource(time.Now().UnixNano())))
	computed, err := crypto.ComputeJoinRequestMIC(*k, pld)
	if err != nil {
		panic(err)
	}
	a.So(checkMIC(*k, append(pld, computed[:]...)), should.BeNil)
	a.So(checkMIC(*k, append(append(pld[1:], pld[0]-1), computed[:]...)), should.NotBeNil)
}
