// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package packetbrokeragent

import (
	"bytes"
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"gopkg.in/square/go-jose.v2"
)

type (
	GatewayUplinkToken gatewayUplinkToken
	AgentUplinkToken   agentUplinkToken
)

func WrapUplinkTokens(gateway, forwarder []byte, agent *AgentUplinkToken) ([]byte, error) {
	return wrapUplinkTokens(gateway, forwarder, (*agentUplinkToken)(agent))
}

func TestWrapGatewayUplinkToken(t *testing.T) {
	a, ctx := test.New(t)
	key := bytes.Repeat([]byte{0x42}, 16)
	encrypter, err := jose.NewEncrypter(jose.A128GCM, jose.Recipient{Algorithm: jose.A128GCMKW, Key: key}, nil)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	wrappedToken, err := wrapGatewayUplinkToken(ctx, &ttnpb.GatewayIdentifiers{GatewayId: "test-gateway"},
		[]byte{0x1, 0x2, 0x3}, encrypter,
	)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	t.Logf("Wrapped token: %q", string(wrappedToken))

	uid, gtwToken, err := unwrapGatewayUplinkToken(wrappedToken, key)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(uid, should.Resemble, unique.ID(ctx, ttnpb.GatewayIdentifiers{GatewayId: "test-gateway"}))
	a.So(gtwToken, should.Resemble, []byte{0x1, 0x2, 0x3})
}

func TestFrequencyPlan(t *testing.T) {
	a, _ := test.New(t)

	fp, err := toPBFrequencyPlan(test.FrequencyPlan("EU_863_870"))
	a.So(err, should.BeNil)
	a.So(fp.LoraSingleSfChannels, should.BeEmpty)
	a.So(fp.FskChannel, should.BeNil)
	a.So(fp.LoraMultiSfChannels, should.HaveLength, 8)
}
